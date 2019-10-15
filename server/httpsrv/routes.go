package httpsrv

import (
	"net/http"

	"github.com/99designs/gqlgen/handler"
	echo "github.com/labstack/echo/v4"

	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/api/graphql/svcgql"

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
	e.GET("/", httputil.InfoHandler(FQN, internal.Version))

	// Register GraphQL routes.
	{
		exec := graphql.NewExecutableSchema(graphql.Config{
			Resolvers: svcgql.NewResolverRoot(
				svcgql.Services{
					Git:          srv.svcs.Git,
					About:        srv.svcs.About,
					Music:        srv.svcs.Music,
					Auth:         srv.svcs.Auth,
					Location:     srv.svcs.Location,
					Scheduling:   srv.svcs.Scheduling,
					Productivity: srv.svcs.Productivity,
				},
				svcgql.Streamers{
					Music: srv.strms.Music,
				},
			),
		})

		e.Any("/graphql", echo.WrapHandler(handler.GraphQL(
			exec,
			handler.ErrorPresenter(gqlutil.PresentError),
			// handler.ComplexityLimit(srv.complexityLimit),
		)))
		e.GET(
			"/playground",
			echo.WrapHandler(handler.Playground("Playground", "/graphql")),
		)
		e.GET(
			"/graphiql",
			echo.WrapHandler(http.HandlerFunc(gqlutil.ServeGraphiQL("/graphql"))),
		)
	}

	return nil
}
