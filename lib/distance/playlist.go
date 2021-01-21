package distance

import (
	"net/url"
	"strconv"
)

// PlaylistState is the state of the server and its playlist.
type PlaylistState struct {
	CurrentLevelIndex int
	Playlist          Playlist
}

// Playlist is the server's playlist.
type Playlist struct {
	Total  int
	Start  int
	Count  int
	Levels []Level
}

// AllPlaylist fetches all playlists and automatically paginates.
func (c *Client) AllPlaylist() (*PlaylistState, error) {
	var state PlaylistState

	for {
		// This loop is overkill, but we're doing it just in case.
		s, err := c.Playlist(state.Playlist.Start, 10240)
		if err != nil {
			return nil, err
		}

		state.CurrentLevelIndex = s.CurrentLevelIndex
		state.Playlist.Levels = append(state.Playlist.Levels, s.Playlist.Levels...)

		last := state.Playlist.Levels[len(state.Playlist.Levels)-1]
		state.Playlist.Start = last.Index + 1

		if state.Playlist.Start >= s.Playlist.Total {
			break
		}
	}

	state.Playlist.Count = state.Playlist.Start
	state.Playlist.Total = state.Playlist.Start
	state.Playlist.Start = 0

	return &state, nil
}

// Playlist fetches a single playlist page.
func (c *Client) Playlist(start, count int) (*PlaylistState, error) {
	v := url.Values{
		"Start": {strconv.Itoa(start)},
		"Count": {strconv.Itoa(count)},
	}

	var p *PlaylistState
	return p, c.getJSON(url.URL{Path: "/playlist", RawQuery: v.Encode()}, &p)
}
