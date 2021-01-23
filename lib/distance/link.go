package distance

import (
	"errors"
	"log"
	"net/http"
	"net/url"
)

// Link links the given player GUID with a private Distance token.
func (c *Client) Link(playerGUID, privToken string) (string, error) {
	type link struct {
		GUID string `json:"Guid"`
	}

	r, err := c.doJSON("POST",
		url.URL{Path: "/link"},
		link{playerGUID},
		http.Header{"Authorization": {"Bearer " + privToken}},
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

// Chat sends a message using the given player token obtained from calling
// Link() and the private Distance token.
func (c *Client) Chat(playerToken, privToken, message string) error {
	type chat struct {
		Message string
	}

	cookie := http.Cookie{
		Name:  "DistanceSession",
		Value: playerToken,
	}

	r, err := c.doJSON("POST",
		url.URL{Path: "/chat"},
		chat{message},
		http.Header{
			"Authorization": {"Bearer " + privToken},
			"Cookie":        {cookie.String()},
		},
	)
	if err != nil {
		return err
	}

	r.Body.Close()

	return nil
}
