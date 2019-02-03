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
		ess.Die("Loading Viper config:", err)
	}
	p, err := mongo.NewProviderUsing(v)
	if err != nil {
		ess.Die("Creating Provider:", err)
	}
	if err = p.Open(); err != nil {
		ess.Die("Opening Provider:", err)
	}

	// Create mood.
	mood := &api.Mood{
		ExtID:   6969,
		Valence: 3,
	}
	if err = p.CreateMood(mood); err != nil {
		ess.Die("Creating mood:", err)
	}
	fmt.Printf("Created mood: %+v", mood)

	// Get mood.
	mood, err = p.GetMood(mood.ID)
	if err != nil {
		ess.Die("Getting mood:", err)
	}
	fmt.Printf("Got mood: %+v\n", mood)
}
