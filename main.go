package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/diamondburned/distant-front/lib/workshop"
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
	state := obs.State()

	level := state.Playlist.Playlist.Levels[0]

	workshopFile, err := workshop.GetFile(level.WorkshopFileID)
	if err != nil {
		log.Fatalln("failed to get workshop file:", err)
	}

	fmt.Println("workshop file URL:", workshopFile.SizedImageURL(128))

	time.Sleep(time.Minute)
}
