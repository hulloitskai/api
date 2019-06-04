package handler

import (
	"net/http"
	"strconv"

	"github.com/stevenxie/api/internal/httputil"

	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
)

// RecentCommitsHandler handles requests for recent commits that I've made.
func RecentCommitsHandler(
	svc api.GitCommitsService,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit := 5
		if qlim := c.QueryParam("limit"); qlim != "" {
			var err error
			if limit, err = strconv.Atoi(qlim); err != nil {
				httputil.SetEchoStatusCode(c, http.StatusBadRequest)
				return errors.Errorf("parsing 'limit' parameter", err)
			}
		}

		// Get recent commits.
		commits, err := svc.RecentGitCommits(limit)
		if err != nil {
			log.WithError(err).Error("Failed to get recent commits.")
			return errors.Errorf("getting recent commits: %w", err)
		}

		return jsonPretty(c, commits)
	}
}
