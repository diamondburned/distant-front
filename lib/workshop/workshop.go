package workshop

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
)

// Client is the default client to use for crawling.
var Client = http.Client{
	Timeout: 8 * time.Second,
}

// File is a workshop shared file.
type File struct {
	Title string
	Image string
}

// SizedImageURL returns the image URL with the given size. It returns the
// original URL if it fails to do so.
func (f File) SizedImageURL(size int) string {
	u, err := url.Parse(f.Image)
	if err != nil {
		return f.Image
	}

	v := u.Query()
	v.Set("imw", strconv.Itoa(size))
	u.RawQuery = v.Encode()

	return u.String()
}

var (
	fileTitle = cascadia.MustCompile("head > title")
	fileImage = cascadia.MustCompile("head > link[rel='image_src']")
)

// CrawlFile crawls the given reader for File.
func CrawlFile(r io.Reader) (*File, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new document")
	}

	var f File
	if title := doc.FindMatcher(fileTitle); title != nil {
		f.Title = strings.TrimPrefix(title.Text(), "Steam Workshop::")
	}

	if f.Title == "" {
		return nil, errors.New("title not found")
	}

	if link := doc.FindMatcher(fileImage); link != nil {
		f.Image = link.AttrOr("href", "")
	}

	return &f, nil
}

// ErrFileNotFound is returned instead of 404.
var ErrFileNotFound = errors.New("workshop file not found")

// GetFile gets a single file with the given ID.
func GetFile(id string) (*File, error) {
	v := url.Values{"id": {id}}

	r, err := Client.Get("https://steamcommunity.com/sharedfiles/filedetails/?" + v.Encode())
	if err != nil {
		return nil, errors.Wrap(err, "failed to reach Steam")
	}
	defer r.Body.Close()

	if r.StatusCode < 200 || r.StatusCode > 299 {
		if r.StatusCode == 404 {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("unexpected status code %d", r.StatusCode)
	}

	return CrawlFile(r.Body)
}
