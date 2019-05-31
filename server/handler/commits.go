package handler

import (
	"net/http"
	"strconv"

	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stevenxie/api/pkg/git"
)

// RecentCommitsHandler handles requests for recent commits that I've made.
func RecentCommitsHandler(
	l zerolog.Logger,
	svc git.RecentCommitsService,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit := 5
		if qlim := c.QueryParam("limit"); qlim != "" {
			var err error
			if limit, err = strconv.Atoi(qlim); err != nil {
				setRequestStatusCode(c, http.StatusBadRequest)
				return errors.Errorf("parsing 'limit' parameter", err)
			}
		}

		// Get recent commits.
		commits, err := svc.RecentCommits(limit)
		if err != nil {
			l.Err(err).Msg("Failed to get recent commits.")
			return errors.Errorf("getting recent commits: %w", err)
		}

		return jsonPretty(c, commits)
	}
}
