package chat

import (
	"io"
	"net/http"

	"github.com/diamondburned/distant-front/internal/frontend"
)

var authenticate = frontend.Templater.Register("authenticate", "index/chat/authenticate.html")

func renderAuth(w http.ResponseWriter, r *http.Request) {
	// Clear cookies.
	writeCookies(w, map[string]string{
		"DistanceSession": "",
		"PrivateToken":    "",
	})

	frontend.ExecuteTemplate(w, r, authenticate)
}

func postAuth(w http.ResponseWriter, r *http.Request) {
	var (
		playerGUID   = r.FormValue("player_guid")
		privateToken = r.FormValue("private_token")
	)

	if playerGUID == "" || privateToken == "" {
		w.WriteHeader(400)
		io.WriteString(w, "missing ?player_guid or ?private_token")
		return
	}

	rs := frontend.GetRenderState(r.Context())

	session, err := rs.Client.Link(playerGUID, privateToken)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, "blame: "+err.Error())
		return
	}

	writeCookies(w, map[string]string{
		"DistanceSession": session,
		"PrivateToken":    privateToken,
	})

	http.Redirect(w, r, "/chat", http.StatusFound)
}

func getCookies(r *http.Request, cookies map[string]string) bool {
	for k := range cookies {
		cookie, err := r.Cookie(k)
		if err != nil {
			return false
		}

		cookies[k] = cookie.Value
	}

	return true
}

func writeCookies(w http.ResponseWriter, cookies map[string]string) {
	for k, v := range cookies {
		http.SetCookie(w, &http.Cookie{
			Name:  k,
			Value: v,
			Path:  "/",
		})
	}
}
