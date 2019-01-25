package main

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/stevenxie/api/pkg/data/airtable"
	ess "github.com/unixpickle/essentials"

	_ "github.com/joho/godotenv/autoload" // load environment from .env
)

const baseID = "appDm0Ts20KbQ0rG0"

func buildConfig() *viper.Viper {
	v := viper.New()
	v.Set("airtable.base_id", baseID)
	return v
}

func main() {
	var (
		v        = buildConfig()
		cfg, err = airtable.ConfigFromViper(v)
	)
	if err != nil {
		ess.Die("Reading config:", err)
	}

	client := airtable.New(cfg)
	moods, err := client.Moods(10)
	if err != nil {
		ess.Die("Fetching mood records:", err)
	}

	fmt.Println("Last 10 moods:")
	for _, mood := range moods {
		fmt.Printf("  • %+v\n", mood)
	}
}
