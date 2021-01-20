package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/diamondburned/distant-front/internal/frontend"
	"github.com/diamondburned/distant-front/internal/server"
	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("failed to load .env:", err)
	}

	c, err := distance.NewClient(os.Getenv("DISTANCE_ENDPOINT"))
	if err != nil {
		log.Fatalln("failed to crete Distance client:", err)
	}

	obs := distance.NewObserver(c, time.Second)

	h := server.New(frontend.RenderState{
		Client:   c,
		Observer: obs,
	})

	log.Println("Listen and serve at :8081")
	log.Fatalln(http.ListenAndServe(":8081", h))
}
