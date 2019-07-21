package main

import (
	"time"

	"github.com/alecthomas/repr"
	"github.com/stevenxie/api/service/location/gmaps"
	ess "github.com/unixpickle/essentials"

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
