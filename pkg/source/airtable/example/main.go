package main

import (
	"fmt"

	"github.com/caarlos0/env"

	"github.com/stevenxie/api/pkg/source/airtable"
	ess "github.com/unixpickle/essentials"

	_ "github.com/joho/godotenv/autoload" // load environment from .env
)

func main() {
	cfg := new(airtable.Config)
	if err := env.Parse(cfg); err != nil {
		ess.Die("Reading config:", err)
	}

	var (
		client     = airtable.New(*cfg)
		moods, err = client.Moods(10)
	)
	if err != nil {
		ess.Die("Fetching mood records:", err)
	}

	fmt.Println("Last 10 moods:")
	for _, mood := range moods {
		fmt.Printf("  • %+v\n", mood)
	}
}
