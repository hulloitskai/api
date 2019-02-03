package main

import (
	"fmt"

	"github.com/stevenxie/api/internal/util"
	"github.com/stevenxie/api/work/airtable"
	ess "github.com/unixpickle/essentials"
)

func main() {
	// Configure MongoDB.
	v, err := util.LoadLocalViper("config")
	if err != nil {
		ess.Die("Loading Viper config:", err)
	}

	client, err := airtable.NewUsing(v)
	if err != nil {
		ess.Die("Creating Airtable client:", err)
	}

	moods, err := client.GetNewMoods()
	if err != nil {
		ess.Die("Fetching mood records:", err)
	}

	fmt.Printf("Got %d moods from Airtable:\n", len(moods))
	for _, mood := range moods {
		fmt.Printf("  • %+v\n", mood)
	}
}
