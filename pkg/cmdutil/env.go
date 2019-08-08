package cmdutil

import (
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/joho/godotenv"
)

// PrepareEnv loads envvars from .env files.
func PrepareEnv() {
	files := []string{".env", ".env.local"}
	for _, file := range files {
		if err := godotenv.Load(file); err != nil {
			// Check to see if error is a not-exists eror.
			{
				pathErr := new(os.PathError)
				if errors.As(err, &pathErr) && os.IsNotExist(pathErr) {
					continue
				}
			}

			// Report error and exit with error code.
			fmt.Fprintf(
				os.Stderr,
				"Failed to read envvars from '%s': %+v\n",
				file, err,
			)
			os.Exit(1)
		}
	}
}
