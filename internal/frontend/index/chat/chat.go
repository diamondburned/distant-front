package chat

import (
	"net/http"

	"github.com/go-chi/chi"
)

func Mount() http.Handler {
	r := chi.NewRouter()
	return r
}
