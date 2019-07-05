package main

import (
	"time"

	ess "github.com/unixpickle/essentials"

	"github.com/alecthomas/repr"
	"github.com/stevenxie/api/provider/google/maps"

	// Automatically load environment variables from '.env' file.
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	h, err := maps.NewHistorian()
	if err != nil {
		ess.Die("Creating historian:", err)
	}

	placemarks, err := h.LocationHistory(time.Now())
	if err != nil {
		ess.Die("Fetching location history:", err)
	}

	repr.Print(placemarks)
}
