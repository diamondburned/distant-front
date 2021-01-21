package workshopimg

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/dgraph-io/badger"
	"github.com/diamondburned/distant-front/internal/httperr"
	"github.com/diamondburned/distant-front/lib/workshop"
	"github.com/go-chi/chi"
)

// CacheOpts is the options for the workshop cache.
type CacheOpts struct {
	// CachePath is the path to the cache database.
	CachePath string
}

// NoCache disables caching.
var NoCache = CacheOpts{}

// Mount mounts the workshop image route.
func Mount(cacheOpts CacheOpts) (http.Handler, error) {
	var c cache

	r := chi.NewRouter()
	r.Get("/{workshopID}", c.redirectImage)

	if cacheOpts.CachePath != "" {
		if err := c.open(cacheOpts.CachePath); err != nil {
			return r, err
		}
	}

	return r, nil
}

// ErrNotFound is used when a workshop file is not found.
var ErrNotFound = httperr.New(404, "file not found")

func verifyID(id string) bool {
	// max u32
	_, err := strconv.ParseUint(id, 10, 32)
	return err == nil
}

type cache struct {
	db *badger.DB
}

func (c *cache) open(path string) (err error) {
	opts := badger.LSMOnlyOptions(path)
	opts.EventLogging = false
	opts.SyncWrites = false
	opts.Truncate = true

	c.db, err = badger.Open(opts)
	return
}

func (c *cache) redirectImage(w http.ResponseWriter, r *http.Request) {
	url, err := c.getFile(chi.URLParam(r, "workshopID"))
	if err != nil {
		httperr.WriteErr(w, err)
		io.WriteString(w, err.Error())
		return
	}

	size, err := strconv.Atoi(r.FormValue("size"))
	if err == nil {
		url = workshop.SizedImageURL(url, size)
	}

	// 301 for caching.
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func (c *cache) getFile(workshopID string) (string, error) {
	if !verifyID(workshopID) {
		return "", httperr.New(400, "invalid ID")
	}

	if c.db == nil {
		return c.getFileDirect(workshopID)
	}

	var imageURL string
	var found bool

	c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(workshopID))
		if err != nil {
			return err
		}
		return item.Value(func(url []byte) error {
			imageURL = string(url)
			found = true
			return nil
		})
	})

	if found {
		if imageURL != "" {
			return imageURL, nil
		}
		return "", ErrNotFound
	}

	v, err := c.getFileDirect(workshopID)
	if err != nil {
		return "", err
	}

	// best effort update - skip error
	c.db.Update(func(txn *badger.Txn) error {
		txn.Set([]byte(workshopID), []byte(v))
		return nil
	})

	return v, nil
}

func (c *cache) getFileDirect(workshopID string) (string, error) {
	f, err := workshop.GetFile(workshopID)
	if err == nil {
		return f.Image, nil
	}

	if errors.Is(err, workshop.ErrFileNotFound) {
		return "", ErrNotFound
	}

	return "", httperr.Wrap(err, 500, "Steam Workshop error")
}
