package main

import (
	"fmt"

	"github.com/stevenxie/api/internal/util"

	"github.com/stevenxie/api"
	"github.com/stevenxie/api/data/mongo"
	ess "github.com/unixpickle/essentials"
)

func main() {
	// Configure MongoDB.
	v, err := util.LoadLocalViper("config")
	if err != nil {
		ess.Die("Loading config:", err)
	}
	p, err := mongo.NewProviderFromViper(v)
	if err != nil {
		ess.Die("Creating provider:", err)
	}
	if err = p.Open(); err != nil {
		ess.Die("Opening provider:", err)
	}

	// Create mood service.
	service := p.MoodService()

	// Create mood.
	mood := &api.Mood{
		ExtID:   6969,
		Valence: 3,
	}
	if err = service.CreateMood(mood); err != nil {
		ess.Die("Creating mood:", err)
	}
	fmt.Printf("Created mood: %+v\n", mood)

	// Get mood.
	mood, err = service.GetMood(mood.ID)
	if err != nil {
		ess.Die("Getting mood:", err)
	}
	fmt.Printf("Got mood: %+v\n", mood)

	// Delete mood.
	err = service.DeleteMood(mood.ID)
	if err != nil {
		ess.Die("Deleting mood:", err)
	}
	fmt.Printf("Deleted mood with ID: %s\n", mood.ID)
}
