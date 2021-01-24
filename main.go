package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/diamondburned/distant-front/internal/frontend"
	"github.com/diamondburned/distant-front/internal/frontend/index"
	"github.com/diamondburned/distant-front/internal/workshopimg"
	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/diamondburned/distant-front/lib/distance/markup"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("failed to load .env:", err)
	}

	var (
		endpoint   = os.Getenv("DISTANCE_ENDPOINT")
		siteName   = os.Getenv("DISTANCE_NAME")
		privToken  = os.Getenv("DISTANCE_PRIVTOKEN")
		listenAddr = os.Getenv("DISTANCE_LISTEN")
	)

	// Make all colors darker.
	markup.ColorModifier = markup.Darken(+0.8, -0.2)

	distanceURL, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalln("invalid $DISTANCE_ENDPOINT:", err)
	}

	c, err := distance.NewClient(distanceURL.String())
	if err != nil {
		log.Fatalln("failed to crete Distance client:", err)
	}

	if privToken != "" {
		c.SetPrivateToken(privToken)
	}

	imgRoute, err := workshopimg.Mount(workshopimg.CacheOpts{
		CachePath: filepath.Join(os.TempDir(), "workshopimg.cache"),
	})
	if err != nil {
		log.Println("Warning: workshop cache load error:", err)
	}

	rs := frontend.RenderState{
		Client:      c,
		Observer:    distance.NewObserver(c, 500*time.Millisecond),
		SiteName:    siteName,
		DistanceURL: distanceURL,
	}

	r := chi.NewRouter()
	r.Mount("/workshopimg", imgRoute)
	r.Mount("/static", frontend.MountStatic())
	r.Mount("/", index.Mount(rs))

	if listenAddr == "" {
		listenAddr = ":8081"
	}

	log.Println("Listen and serve at", listenAddr)
	log.Fatalln(http.ListenAndServe(listenAddr, r))
}
