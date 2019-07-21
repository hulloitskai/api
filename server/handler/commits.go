package handler

import (
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/stevenxie/api/pkg/httputil"
	cm "github.com/stevenxie/api/service/commits"
)

// RecentCommitsHandler handles requests for recent commits that I've made.
func RecentCommitsHandler(
	svc cm.Service,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit := 5
		const limitParamName = "limit"
		if qlim := c.QueryParam(limitParamName); qlim != "" {
			var err error
			if limit, err = strconv.Atoi(qlim); err != nil {
				httputil.SetEchoStatusCode(c, http.StatusBadRequest)
				return errors.Wrapf(err, "bad parameter '%s'", limitParamName)
			}
		}

		// Get recent commits.
		commits, err := svc.RecentCommits(limit)
		if err != nil {
			log.WithError(err).Error("Failed to get recent commits.")
			return errors.Wrap(err, "getting recent commits")
		}
		return jsonPretty(c, commits)
	}
}
