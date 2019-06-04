package zero

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

// Initialize noopLogger.
func init() {
	noopLogger = logrus.New()
	noopLogger.SetLevel(logrus.PanicLevel)
	noopLogger.SetOutput(ioutil.Discard)
}

var noopLogger *logrus.Logger

// Logger returns a no-op logrus.Logger.
func Logger() *logrus.Logger { return noopLogger }
