package index

import (
	"net/http"
	"sort"

	"github.com/diamondburned/distant-front/internal/frontend"
	"github.com/diamondburned/distant-front/internal/frontend/index/chat"
	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/go-chi/chi"
)

var index = frontend.Templater.Register("index", "index/index.html")

func init() {
	frontend.Templater.Func("sortPlayers", sortPlayers)
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

func Mount(rs frontend.RenderState) http.Handler {
	// force preload template for early error catching
	frontend.Templater.Preload()

	r := chi.NewRouter()
	r.Use(frontend.InjectRenderState(rs))

	r.Mount("/chat", chat.Mount())
	r.Get("/", renderIndex)

	return r
}

func renderIndex(w http.ResponseWriter, r *http.Request) {
	frontend.ExecuteTemplate(w, r, index)
}
