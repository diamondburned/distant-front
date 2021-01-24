package chat

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/diamondburned/distant-front/internal/frontend"
	"github.com/diamondburned/distant-front/internal/frontend/index/link"
	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/go-chi/chi"
)

var chat = frontend.Templater.Register("chat", "index/chat/chat.html")

func init() {
	frontend.Templater.Func("reverseMessages", reverseMessages)
	frontend.Templater.Func("htmlTime", func(t time.Time) string {
		return t.Format("2006-01-02T15:04")
	})
}

// Mount mounts the chat routes.
func Mount() http.Handler {
	r := chi.NewRouter()
	r.Get("/", render)
	r.Post("/", sendMessage)

	r.Post("/unlink", unlinkSession)
	r.Get("/listen/{afterID}", listen)

	return r
}

func unlinkSession(w http.ResponseWriter, r *http.Request) {
	link.ClearDistanceSession(w)
	http.Redirect(w, r, "/chat", http.StatusFound)
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	session := link.GetDistanceSession(r)
	if session == "" {
		w.WriteHeader(401)
		io.WriteString(w, "action not permitted: missing session")
		return
	}

	message := r.FormValue("m")
	rs := frontend.GetRenderState(r.Context())

	if err := rs.Client.Chat(session, message); err != nil {
		w.WriteHeader(500)
		io.WriteString(w, "failed to send message: "+err.Error())
		return
	}

	http.Redirect(w, r, "/chat", http.StatusFound)
}

type renderData struct {
	frontend.RenderState
	IsLinked bool
}

func render(w http.ResponseWriter, r *http.Request) {
	data := renderData{
		RenderState: frontend.GetRenderState(r.Context()),
		IsLinked:    link.GetDistanceSession(r) != "",
	}

	if err := chat.Execute(w, data); err != nil {
		log.Println("Error rendering:", err)
	}
}

const (
	// MagicExpire is the magical expired HTML string.
	MagicExpire = "<!-- SESSION EXPIRED -->"
	// MagicHalted is written when the server is halted.
	MagicHalted = "<!-- SERVER HALTED -->"
)

func listen(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	id := chi.URLParam(r, "afterID")
	rs := frontend.GetRenderState(r.Context())

	// Try and fetch the player's GUID so we can keep track of it.
	var playerGUID string

	if session := link.GetDistanceSession(r); session != "" {
		g, err := rs.Client.PlayerGUID(session)
		if err != nil {
			w.WriteHeader(401)
			io.WriteString(w, "invalid session token")
			return
		}
		playerGUID = g
	}

	flusher, canFlush := w.(http.Flusher)

	evCh, cancel := rs.Observer.Subscribe()
	defer cancel()

	for {
		select {
		case <-r.Context().Done():
			// Request cancelled; bail with OK.
			return

		case _, ok := <-evCh:
			if !ok {
				// Observer is halted; bail with error.
				io.WriteString(w, MagicHalted)
				return
			}

			// Get the latest event manually in case our event loop is lagged
			// behind.
			ev := rs.Observer.State()

			// If there are no messages for some reason, then keep waiting.
			if len(ev.Summary.ChatLog) == 0 {
				continue
			}
			// If the last message is still the currently waiting message, then
			// keep waiting. Else, immediately update the latest ID.
			if ev.Summary.ChatLog[len(ev.Summary.ChatLog)-1].GUID == id {
				continue
			}

			// Find the index to send.
			ix := lookBackwards(ev.Summary.ChatLog, id)
			for i := len(ev.Summary.ChatLog) - 1; i != ix; i-- {
				frontend.Templater.Execute(w, "chat-message", ev.Summary.ChatLog[i])
				// Delimit using a NULL byte. By using a proper delimiter, we
				// don't need to worry about properly handling HTTP flushes.
				w.Write([]byte{0})
			}

			// Update the latest ID.
			id = ev.Summary.ChatLog[len(ev.Summary.ChatLog)-1].GUID

			// Optionally flush the events over.
			if canFlush {
				flusher.Flush()
			}

			// Confirm that the player still has the same GUID. Drop as soon as
			// that is false.
			if playerGUID != "" && ev.Summary.FindPlayer(playerGUID) == nil {
				// Write a special constant to trigger the frontend and bail.
				io.WriteString(w, MagicExpire)
				w.Write([]byte{0})
				return
			}
		}
	}
}

func reverseMessages(msgs []distance.ChatMessage) []distance.ChatMessage {
	msgs = append([]distance.ChatMessage(nil), msgs...)

	for i := len(msgs)/2 - 1; i >= 0; i-- {
		opp := len(msgs) - 1 - i
		msgs[i], msgs[opp] = msgs[opp], msgs[i]
	}

	return msgs
}

// lookBackwards looks backwards in the given slice for the message with the
// given ID. The returned integer is the index of that message; 0 is returned
// if nothing is found.
func lookBackwards(msgs []distance.ChatMessage, id string) int {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].GUID == id {
			return i
		}
	}
	return 0
}
