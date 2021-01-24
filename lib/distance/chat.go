package distance

import (
	"net/http"
	"net/url"
)

// Chat sends a message using the given player token obtained from calling
// Link() and the private Distance token.
func (c *Client) Chat(playerToken, message string) error {
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
		http.Header{"Cookie": {cookie.String()}},
	)
	if err != nil {
		return err
	}

	r.Body.Close()

	return nil
}

// ServerChat sends a message as the server using the private Distance token.
func (c *Client) ServerChat(message string) error {
	type chat struct {
		Message string
	}

	r, err := c.doJSON("POST",
		url.URL{Path: "/serverchat"},
		chat{message},
		nil,
	)
	if err != nil {
		return err
	}

	r.Body.Close()

	return nil
}
