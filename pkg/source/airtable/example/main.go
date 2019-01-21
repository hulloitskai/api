package main

import (
	"fmt"

	"github.com/stevenxie/api/pkg/airtable"
	ess "github.com/unixpickle/essentials"
)

func main() {
	cfg, err := readConfig()
	if err != nil {
		ess.Die("Reading config:", err)
	}

	client, err := airtable.New(cfg)
	if err != nil {
		ess.Die("Creating Airtable client:", err)
	}

	moods, err := client.Moods(10)
	if err != nil {
		ess.Die("Fetching mood records:", err)
	}

	fmt.Println(moods)
}
