package cli

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/joho/godotenv"
	"github.com/stevenxie/api/internal/info"
	ess "github.com/unixpickle/essentials"
)

func loadEnv() {
	var (
		sysEnvPath = fmt.Sprintf("/etc/%s/.env", info.Namespace)
		err        = godotenv.Load(".env", ".env.local", sysEnvPath)
	)

	if (err != nil) &&
		!strings.Contains(err.Error(), "no such file or directory") {
		ess.Die("Error while reading .env file:", err)
	}
}

func buildLogger() (*zap.SugaredLogger, error) {
	var (
		raw *zap.Logger
		err error
	)
	if os.Getenv("GOENV") == "development" {
		raw, err = zap.NewDevelopment()
	} else {
		cfg := zap.NewProductionConfig()
		cfg.Encoding = "console"
		raw, err = cfg.Build()
	}
	if err != nil {
		return nil, err
	}
	return raw.Sugar(), nil
}
