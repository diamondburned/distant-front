package chat

import (
	"io"
	"net/http"
	"time"

	"github.com/diamondburned/distant-front/internal/frontend"
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
	r.Get("/after/{afterID}", renderAfter)
	return r
}

const (
	// MinWait is the minimum time to wait before yielding the request back.
	// This is done to prevent overloading the server.
	MinWait = 500 * time.Millisecond
	// MaxWait is the maximum wait time before the request returns.
	MaxWait = 30 * time.Second
)

func render(w http.ResponseWriter, r *http.Request) {
	frontend.ExecuteTemplate(w, r, chat)
}

func renderAfter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "afterID")
	rs := frontend.GetRenderState(r.Context())

	timeout := time.NewTimer(MaxWait)
	defer timeout.Stop()

	evCh, cancel := rs.Observer.Subscribe()
	defer cancel()

	for {
		select {
		case <-r.Context().Done():
			// Request cancelled; bail with OK.
			return
		case <-timeout.C:
			// Timeout; bail with OK.
			return

		case ev, ok := <-evCh:
			if !ok {
				// Observer is halted; bail with error.
				w.WriteHeader(500)
				io.WriteString(w, "Server halted.")
				return
			}

			// If there are no messages for some reason, then keep waiting.
			if len(ev.Summary.ChatLog) == 0 {
				continue
			}
			// If the last message is still the currently waiting message, then
			// keep waiting.
			if ev.Summary.ChatLog[len(ev.Summary.ChatLog)-1].GUID == id {
				continue
			}

			// Find the index to send.
			ix := lookBackwards(ev.Summary.ChatLog, id)
			for i := len(ev.Summary.ChatLog) - 1; i != ix; i-- {
				frontend.Templater.Execute(w, "chat-message", ev.Summary.ChatLog[i])
			}

			return
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
