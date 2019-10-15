package main

import (
	"os"

	"github.com/dmksnnk/sentryhook"
	raven "github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/internal"
	"go.stevenxie.me/gopkg/cmdutil"
)

func buildRaven() *raven.Client {
	rc, err := raven.New(os.Getenv("SENTRY_DSN"))
	if err != nil {
		cmdutil.Fatalf("Failed to build Raven client: %v\n", err)
	}

	// Configure client.
	if env := os.Getenv("GOENV"); env != "" {
		rc.SetEnvironment(env)
	}
	rc.SetRelease(internal.Version)

	return rc
}

// buildLogger builds an application-level logger, which also captures errors
// using Sentry.
func buildLogger(rc *raven.Client) *logrus.Entry {
	log := logrus.New()
	log.SetOutput(os.Stdout)

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
	// Integrate error reporting with Sentry.
	hook := sentryhook.New(rc)
	hook.SetAsync(logrus.ErrorLevel)
	hook.SetSync(logrus.PanicLevel, logrus.FatalLevel)
	log.AddHook(hook)

	return logrus.NewEntry(log)
}
