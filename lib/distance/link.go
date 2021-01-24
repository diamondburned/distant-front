package distance

import (
	"encoding/json"
	"log"
	"net/url"

	"github.com/pkg/errors"
)

// Link links the given player GUID with a private Distance token.
func (c *Client) Link(playerGUID string) (string, error) {
	type link struct {
		GUID string `json:"Guid"`
	}

	r, err := c.doJSON("POST",
		url.URL{Path: "/link"},
		link{playerGUID},
		nil,
	)
	if err != nil {
		return "", err
	}
	r.Body.Close()

	for _, cookie := range r.Cookies() {
		if cookie.Name == "DistanceSession" {
			log.Printf("%#v\n", cookie)
			return cookie.Value, nil
		}
	}

	return "", errors.New("cookie not received; server error")
}

// Links is the links registry linking players and API sessions.
type Links struct {
	CodesForward map[string]string // 6-digit link code, UnityPlayer GUID
	CodesReverse map[string]string // 6-digit link code, Session token
	Links        map[string]string // Session token, UnityPlayer GUID
}

// Links fetches the links registry from the server.
func (c *Client) Links() (*Links, error) {
	r, err := c.doJSON("GET",
		url.URL{Path: "/links"},
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var links Links
	if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
		return nil, errors.Wrap(err, "failed to decode links body")
	}

	return &links, nil
}

// ErrLinkCodeNotFound is returned when the link code cannot be found. It is
// returned by LinkSession.
var ErrLinkCodeNotFound = errors.New("link code not found")

// LinkSession matches the given link code to a player's GUID and links that
// GUID to a new session. The session token is returned.
func (c *Client) LinkSession(linkCode string) (string, error) {
	links, err := c.Links()
	if err != nil {
		return "", errors.Wrap(err, "failed to get links")
	}

	guid, ok := links.CodesForward[linkCode]
	if !ok {
		return "", ErrLinkCodeNotFound
	}

	session, err := c.Link(guid)
	if err != nil {
		return "", errors.Wrap(err, "failed to link")
	}

	return session, nil
}
