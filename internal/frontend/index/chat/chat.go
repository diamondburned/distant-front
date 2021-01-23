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
	r.Use(noSniff)
	r.Get("/", render)
	r.Get("/listen/{afterID}", listen)
	return r
}

func noSniff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func render(w http.ResponseWriter, r *http.Request) {
	frontend.ExecuteTemplate(w, r, chat)
}

func listen(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "afterID")
	rs := frontend.GetRenderState(r.Context())

	flusher, canFlush := w.(http.Flusher)

	evCh, cancel := rs.Observer.Subscribe()
	defer cancel()

	for {
		select {
		case <-r.Context().Done():
			// Request cancelled; bail with OK.
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
			// keep waiting. Else, immediately update the latest ID.
			if last := ev.Summary.ChatLog[len(ev.Summary.ChatLog)-1]; last.GUID == id {
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
