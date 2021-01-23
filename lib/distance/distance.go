package distance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Client is a Distance Server client with a custom endpoint.
type Client struct {
	Client    http.Client
	endpoint  url.URL
	privToken string

	// internal state
	ctx context.Context
}

// NewClient creates a new Distance Server client.
func NewClient(endpoint, privToken string) (*Client, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "invalid endpoint URL")
	}

	return &Client{
		Client:    http.Client{Timeout: 10 * time.Second},
		endpoint:  *url,
		privToken: privToken,
		ctx:       context.Background(),
	}, nil
}

// WithContext creates a copy of the Client with the given context for
// cancelation and timeout.
func (c *Client) WithContext(ctx context.Context) *Client {
	cpy := *c
	cpy.ctx = ctx
	return &cpy
}

func (c *Client) getJSON(u url.URL, dst interface{}) error {
	r, err := c.doJSON("GET", u, dst, nil)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return errors.Wrap(err, "failed to decode JSON body")
	}

	return nil
}

func (c *Client) doJSON(
	method string, u url.URL, in interface{}, h http.Header) (*http.Response, error) {

	u.Scheme = c.endpoint.Scheme
	u.Host = c.endpoint.Host

	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal body")
		}
		body = bytes.NewReader(b)
	}

	rq, err := http.NewRequestWithContext(c.ctx, method, u.String(), body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	if h != nil {
		rq.Header = h
	}

	// This is safe to do just because we're not passing raw data to the
	// frontend.
	if c.privToken != "" {
		rq.Header.Set("Authorization", "Bearer "+c.privToken)
	}

	resp, err := c.Client.Do(rq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return resp, nil
}

// Observer observes the server for changes periodically.
type Observer struct {
	mutex sync.RWMutex
	waitg sync.WaitGroup
	state ObservedState

	subMu sync.Mutex
	subs  map[chan ObservedState]struct{}

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
		sig:    make(chan struct{}, 1), // allow queueing
		done:   make(chan struct{}),
		subs:   map[chan ObservedState]struct{}{},

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

	state := ObservedState{
		LastRenew:     tick,
		PlaylistState: playlist,
		Summary:       summary,
	}

	obs.mutex.Lock()
	obs.state = state
	obs.mutex.Unlock()

	obs.subMu.Lock()
	defer obs.subMu.Unlock()

	for ch := range obs.subs {
		select {
		case ch <- state:
		default:
		}
	}
}

// Subscribe subscribes to the current observer. The returned channel will be
// ticked everytime the observed state is updated. If the returned callback is
// called, the channel will be closed, and the channel will be unsubscribed.
//
// If the observee fail to observe the next tick, then it is queued once. All
// consecutive states will be dropped before the buffer is free again.
//
// When the Observer is shut down, its subscribed channels will be closed.
func (obs *Observer) Subscribe() (<-chan ObservedState, func()) {
	observee := make(chan ObservedState, 1)

	obs.subMu.Lock()
	defer obs.subMu.Unlock()

	// If the observer is invalidated, then return an already closed channel.
	if obs.subs == nil {
		close(observee)
		return observee, func() {}
	}

	obs.subs[observee] = struct{}{}

	return observee, func() {
		obs.subMu.Lock()
		// Check that the subscribed state still exists. We should only delete
		// it if it does.
		if _, ok := obs.subs[observee]; ok {
			delete(obs.subs, observee)
		}
		obs.subMu.Unlock()

		// Only close the observee channel after deleting from the map and
		// unlocking to prevent sending to a closed channel.
		close(observee)
	}
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
	// Mark that the observer should be renewed. If the channel is full, then
	// the observer will eventually be renewed; therefore we don't have to do it
	// again.
	select {
	case obs.sig <- struct{}{}:
	default:
	}
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

	// Close all channels.
	obs.subMu.Lock()
	defer obs.subMu.Unlock()

	for ch := range obs.subs {
		close(ch)
	}

	obs.subs = nil
}
