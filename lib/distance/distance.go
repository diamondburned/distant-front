package distance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Client is a Distance Server client with a custom endpoint.
type Client struct {
	Client   http.Client
	endpoint url.URL

	// internal state
	ctx context.Context
}

// NewClient creates a new Distance Server client.
func NewClient(endpoint string) (*Client, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "invalid endpoint URL")
	}

	return &Client{
		Client:   http.Client{Timeout: 10 * time.Second},
		endpoint: *url,
		ctx:      context.Background(),
	}, nil
}

// WithContext creates a copy of the Client with the given context for
// cancelation and timeout.
func (c *Client) WithContext(ctx context.Context) *Client {
	cpy := *c
	cpy.ctx = ctx
	return &cpy
}

// WithSession creates a copy of the Client with the given context and session
// for individual authentication.
func (c *Client) WithSession(ctx context.Context, session string) *Client {
	cookies, _ := cookiejar.New(nil)
	cookies.SetCookies(&c.endpoint, []*http.Cookie{{
		Path:   "/",
		Name:   "DistanceSession",
		Value:  session,
		Domain: c.endpoint.Hostname(),
	}})

	c = c.WithContext(ctx)
	c.Client.Jar = cookies

	return c
}

func (c *Client) getJSON(u url.URL, dst interface{}) error {
	u.Scheme = c.endpoint.Scheme
	u.Host = c.endpoint.Host

	rq, err := http.NewRequestWithContext(c.ctx, "GET", u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	resp, err := c.Client.Do(rq)
	if err != nil {
		return errors.Wrap(err, "failed to do request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return errors.Wrap(err, "failed to decode JSON body")
	}

	return nil
}

// Observer observes the server for changes periodically.
type Observer struct {
	mutex sync.RWMutex
	waitg sync.WaitGroup
	state ObservedState

	// constants
	client *Client
	sig    chan struct{}
	done   chan struct{}

	// OnError is called on a fetch error. By default, it logs to console.
	OnError func(error)
}

// ObservedState describes the last observed state.
type ObservedState struct {
	Summary       *Summary
	PlaylistState *PlaylistState
	LastRenew     time.Time
}

// NewObserver creates a new periodic observer.
func NewObserver(c *Client, dura time.Duration) *Observer {
	obs := Observer{
		client: c,
		sig:    make(chan struct{}),
		done:   make(chan struct{}),

		OnError: func(err error) {
			log.Println("[distance] Observer error:", err)
		},
	}

	// Synchronously renew the state before starting the loop.
	tick := time.Now()
	obs.refetch(tick)

	go func() {
		defer close(obs.done)

		ticker := time.NewTicker(dura)
		defer ticker.Stop()

		for {
			select {
			case tick = <-ticker.C:
			case _, ok := <-obs.sig:
				if !ok {
					return
				}
				tick = time.Now()
			}
			obs.refetch(tick)
		}
	}()

	return &obs
}

func (obs *Observer) refetch(tick time.Time) {
	obs.waitg.Add(1)

	var playlist *PlaylistState
	go func() {
		p, err := obs.client.AllPlaylist()
		if err != nil {
			obs.OnError(errors.Wrap(err, "failed to get all playlists"))
		}
		playlist = p
		obs.waitg.Done()
	}()

	summary, err := obs.client.Summary()
	if err != nil {
		obs.OnError(errors.Wrap(err, "failed to get summary"))
	}

	obs.waitg.Wait()

	obs.mutex.Lock()
	defer obs.mutex.Unlock()

	obs.state.LastRenew = tick
	obs.state.PlaylistState = playlist
	obs.state.Summary = summary
}

// State returns the current observed state. The user must not mutate fields
// inside the state, as that is racy.
func (obs *Observer) State() ObservedState {
	obs.mutex.RLock()
	defer obs.mutex.RUnlock()

	return obs.state
}

// Renew queues a renew. It does not wait for the renew to finish.
func (obs *Observer) Renew() {
	obs.sig <- struct{}{}
}

// Stop stops the observer. Calling stop more than once does nothing.
func (obs *Observer) Stop() {
	select {
	case <-obs.done:
		return
	default:
	}

	close(obs.sig)
	<-obs.done
}
