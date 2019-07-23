package main

import (
	"time"

	"github.com/alecthomas/repr"
	ess "github.com/unixpickle/essentials"
	"go.stevenxie.me/api/service/location/gmaps"

	// Automatically load environment variables from '.env' file.
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	h, err := gmaps.NewHistorian()
	if err != nil {
		ess.Die("Creating historian:", err)
	}

	placemarks, err := h.LocationHistory(time.Now())
	if err != nil {
		ess.Die("Fetching location history:", err)
	}

	repr.Print(placemarks)
}
