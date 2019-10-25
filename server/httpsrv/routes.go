package httpsrv

import (
	"net/http"
	"os"

	"github.com/99designs/gqlgen/handler"
	echo "github.com/labstack/echo/v4"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/internal"
	"go.stevenxie.me/api/pkg/gqlutil"
	"go.stevenxie.me/api/pkg/httputil"
)

func (srv *Server) registerRoutes() error {
	e := srv.echo
	e.HTTPErrorHandler = httputil.ErrorHandler(
		srv.log.WithField("handler", "error"),
	)

	// Register metadata route.
	e.Match(
		[]string{http.MethodGet, http.MethodHead}, "/",
		httputil.InfoHandler(name.OfType((*Server)(nil)), internal.Version),
	)

	// Add GraphQL and GraphiQL endpoints.
	e.Any("/graphql", echo.WrapHandler(srv.gqlHandler()))
	e.GET(
		"/graphiql",
		echo.WrapHandler(http.HandlerFunc(gqlutil.ServeGraphiQL("./graphql"))),
	)

	// Only enable playground in development.
	if os.Getenv("GOENV") == "development" {
		e.GET(
			"/playground",
			echo.WrapHandler(handler.Playground("Playground", "/graphql")),
		)
	}

	return nil
}
