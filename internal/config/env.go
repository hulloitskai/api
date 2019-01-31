package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/stevenxie/api/internal/info"
)

// LoadDotEnv loads environment variables from '.env' files.
func LoadDotEnv() error {
	var (
		sysEnvPath = fmt.Sprintf("/etc/%s/.env", info.Namespace)
		err        = godotenv.Load(".env", ".env.local", sysEnvPath)
	)
	if (err != nil) &&
		!strings.Contains(err.Error(), "no such file or directory") {
		return err
	}
	return nil
}
