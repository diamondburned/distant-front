package index

import (
	"log"
	"net/http"

	"github.com/diamondburned/distant-front/internal/frontend"
	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/go-chi/chi"
)

var index = frontend.Templater.Register("index", "index/index.html")

func Mount(rs frontend.RenderState) http.Handler {
	// force preload template for early error catching
	frontend.Templater.Preload()

	r := chi.NewRouter()
	r.Use(frontend.InjectRenderState(rs))
	r.Get("/", renderIndex)
	return r
}

type renderData struct {
	distance.ObservedState
}

func renderIndex(w http.ResponseWriter, r *http.Request) {
	if err := index.Execute(w, frontend.GetRenderState(r.Context())); err != nil {
		log.Println("Error rendering:", err)
	}
}
