package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/cmdutil"
)

// Build an application-level logger, which also captures errors using Sentry.
func buildLogger() *logrus.Entry {
	log := logrus.New()
	log.SetOutput(os.Stderr)

	// Set logger level.
	{
		const key = "LOGRUS_LEVEL"
		if lvl, ok := os.LookupEnv(key); ok {
			level, err := logrus.ParseLevel(lvl)
			if err != nil {
				cmdutil.Fatalf("Failed to parse '%s' as a logrus.Level.\n", key)
			}
			log.SetLevel(level)
		} else if os.Getenv("GOENV") == "development" {
			log.SetLevel(logrus.DebugLevel)
		}
	}

	return logrus.NewEntry(log)
}
