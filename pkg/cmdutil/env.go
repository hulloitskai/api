package cmdutil

import (
	"strings"

	"github.com/joho/godotenv"
	ess "github.com/unixpickle/essentials"
)

// PrepareEnv loads envvars from .env files.
func PrepareEnv() {
	if err := godotenv.Load(".env", ".env.local"); err != nil {
		if !strings.Contains( // unknown error
			err.Error(),
			"no such file or directory",
		) {
			ess.Die("Error reading '.env' file:", err)
		}
	}
}
