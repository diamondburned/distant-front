package server

import (
	"net/http"

	"github.com/diamondburned/distant-front/internal/frontend"
	"github.com/diamondburned/distant-front/internal/frontend/index"
	"github.com/go-chi/chi"
)

func New(rs frontend.RenderState) http.Handler {
	r := chi.NewRouter()
	r.Mount("/", index.Mount(rs))
	return r
}
