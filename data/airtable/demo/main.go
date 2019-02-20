package main

import (
	"fmt"

	"github.com/stevenxie/api/data/airtable"
	"github.com/stevenxie/api/internal/util"
	ess "github.com/unixpickle/essentials"
)

func main() {
	// Configure MongoDB.
	v, err := util.LoadLocalViper("config")
	if err != nil {
		ess.Die("Loading Viper config:", err)
	}

	p, err := airtable.NewProviderViper(v)
	if err != nil {
		ess.Die("Creating Airtable client:", err)
	}

	moods, err := p.MoodSource().GetNewMoods()
	if err != nil {
		ess.Die("Fetching mood records:", err)
	}

	fmt.Printf("Got %d moods from Airtable:\n", len(moods))
	for _, mood := range moods {
		fmt.Printf("  • %+v\n", mood)
	}
}
