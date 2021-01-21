package main

import (
	"log"
	"net/http"
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

	// Make all colors darker.
	markup.ColorModifier = markup.Darken(+0.5, -0.25)

	c, err := distance.NewClient(os.Getenv("DISTANCE_ENDPOINT"))
	if err != nil {
		log.Fatalln("failed to crete Distance client:", err)
	}

	imgRoute, err := workshopimg.Mount(workshopimg.CacheOpts{
		CachePath: filepath.Join(os.TempDir(), "workshopimg.cache"),
	})
	if err != nil {
		log.Println("Warning: workshop cache load error:", err)
	}

	rs := frontend.RenderState{
		Client:   c,
		Observer: distance.NewObserver(c, time.Second),
		SiteName: "Distant Front",
	}

	r := chi.NewRouter()
	r.Mount("/workshopimg", imgRoute)
	r.Mount("/static", frontend.MountStatic())
	r.Mount("/", index.Mount(rs))

	log.Println("Listen and serve at :8081")
	log.Fatalln(http.ListenAndServe(":8081", r))
}
