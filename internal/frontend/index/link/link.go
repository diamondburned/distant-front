package link

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/diamondburned/distant-front/internal/frontend"
	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/go-chi/chi"
)

var link = frontend.Templater.Register("link", "index/link/link.html")

// GetDistanceSession gets the Distance session token from the request cookies.
// An empty string is returned if none is found.
func GetDistanceSession(r *http.Request) string {
	cookie, err := r.Cookie("DistanceSession")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func Mount() http.Handler {
	r := chi.NewRouter()
	r.Get("/", renderAuth)
	r.Post("/", postAuth)
	return r
}

type renderAuthData struct {
	frontend.RenderState
	Error    string
	Unlinked bool
}

func renderAuth(w http.ResponseWriter, r *http.Request) {
	// Clear cookies.
	writeDistanceSession(w, "")

	executeAuthenticateTmpl(w, renderAuthData{
		RenderState: frontend.GetRenderState(r.Context()),
		Unlinked:    r.FormValue("unlinked") != "",
	})
}

func executeAuthenticateTmpl(w http.ResponseWriter, data renderAuthData) {
	if err := link.Execute(w, data); err != nil {
		log.Println("Error rendering:", err)
	}
}

func postAuth(w http.ResponseWriter, r *http.Request) {
	linkCode := r.FormValue("link_code")
	rs := frontend.GetRenderState(r.Context())

	if linkCode == "" {
		w.WriteHeader(400)
		executeAuthenticateTmpl(w, renderAuthData{
			RenderState: rs,
			Error:       "missing link code",
		})

		return
	}

	s, err := rs.Client.LinkSession(linkCode)
	if err != nil {
		if errors.Is(err, distance.ErrLinkCodeNotFound) {
			w.WriteHeader(400)
			executeAuthenticateTmpl(w, renderAuthData{
				RenderState: rs,
				Error:       "invalid link code",
			})

			return
		}

		// Treat by default errors received are server-side errors.
		w.WriteHeader(500)
		executeAuthenticateTmpl(w, renderAuthData{
			RenderState: rs,
			Error:       err.Error(),
		})

		return
	}

	writeDistanceSession(w, s)
	http.Redirect(w, r, "/chat", http.StatusFound)
}

func writeDistanceSession(w http.ResponseWriter, session string) {
	cookie := http.Cookie{
		Name:  "DistanceSession",
		Value: session,
		Path:  "/",
	}

	if session == "" {
		cookie.Expires = time.Unix(0, 0)
	}

	http.SetCookie(w, &cookie)
}
