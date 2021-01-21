package index

import (
	"net/http"

	"github.com/diamondburned/distant-front/internal/frontend"
	"github.com/diamondburned/distant-front/internal/frontend/index/chat"
	"github.com/go-chi/chi"
)

var index = frontend.Templater.Register("index", "index/index.html")

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
