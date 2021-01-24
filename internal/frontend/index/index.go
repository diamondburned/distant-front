package index

import (
	"net/http"
	"sort"

	"github.com/diamondburned/distant-front/internal/frontend"
	"github.com/diamondburned/distant-front/internal/frontend/index/chat"
	"github.com/diamondburned/distant-front/internal/frontend/index/link"
	"github.com/diamondburned/distant-front/internal/tmplutil"
	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/go-chi/chi"
)

var index = frontend.Templater.Register("index", "index/index.html")

func init() {
	frontend.Templater.Func("sortPlayers", sortPlayers)
	frontend.Templater.Func("latestMessages", latestMessages)
}

func sortPlayers(players []distance.Player) []distance.Player {
	if len(players) < 2 {
		return players
	}

	players = append([]distance.Player(nil), players...)

	sort.SliceStable(players, func(i, j int) bool {
		// Put the finished players in front.
		if players[i].Car.Finished {
			if !players[j].Car.Finished {
				return true
			}
			// Put players who finished first before.
			return players[i].Car.FinishData < players[j].Car.FinishData
		}

		// Put spectators last.
		return !players[i].Car.Spectator
	})

	return players
}

func latestMessages(msgs []distance.ChatMessage, max int) []distance.ChatMessage {
	var latest = make([]distance.ChatMessage, 0, max)

	for i := len(msgs) - 1; i >= 0 && len(latest) < max; i-- {
		switch msg := msgs[i]; msg.Description {
		case "AutoServer:Tip":
			continue
		default:
			latest = append(latest, msg)
		}
	}

	// We appended the latest ones first, so we have to flip the slice.
	for i := len(latest)/2 - 1; i >= 0; i-- {
		opp := len(latest) - 1 - i
		latest[i], latest[opp] = latest[opp], latest[i]
	}

	return latest
}

func Mount(rs frontend.RenderState) http.Handler {
	// force preload template for early error catching
	frontend.Templater.Preload()

	r := chi.NewRouter()
	r.Use(frontend.InjectRenderState(rs))

	r.Group(func(r chi.Router) {
		r.Mount("/chat", chat.Mount())
		r.Mount("/link", link.Mount())
	})

	r.Group(func(r chi.Router) {
		r.Use(tmplutil.AlwaysFlush)
		r.Get("/body", renderBody)
		r.Get("/", renderIndex)
	})

	return r
}

func noSniff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func renderIndex(w http.ResponseWriter, r *http.Request) {
	frontend.ExecuteTemplate(w, r, index)
}

func renderBody(w http.ResponseWriter, r *http.Request) {
	frontend.ExecuteNamedTemplate(w, r, "index-body")
}
