package gqlsrv

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/websocket"

	"go.stevenxie.me/api/v2/graphql"
	"go.stevenxie.me/api/v2/graphql/svcgql"
	"go.stevenxie.me/api/v2/pkg/gqlutil"
)

func (srv *Server) gqlHandler() http.Handler {
	exec := graphql.NewExecutableSchema(graphql.Config{
		Resolvers: svcgql.NewResolverRoot(
			svcgql.Services{
				Git:          srv.svcs.Git,
				Auth:         srv.svcs.Auth,
				About:        srv.svcs.About,
				Music:        srv.svcs.Music,
				Transit:      srv.svcs.Transit,
				Location:     srv.svcs.Location,
				Scheduling:   srv.svcs.Scheduling,
				Productivity: srv.svcs.Productivity,
			},
			svcgql.Streamers{
				Music: srv.strms.Music,
			},
		),
	})

	// Configure GraphQL handler.
	handlerOpts := []handler.Option{
		handler.ErrorPresenter(gqlutil.PresentError),
		handler.WebsocketKeepAliveDuration(10 * time.Second),
		handler.WebsocketUpgrader(websocket.Upgrader{
			// Allow access from all origins.
			CheckOrigin:     func(*http.Request) bool { return true },
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}),
	}
	if s := srv.sentry; s != nil {
		handlerOpts = append(
			handlerOpts,
			handler.RecoverFunc(gqlutil.SentryRecoverFunc(s, srv.log)),
		)
	}

	return handler.GraphQL(exec, handlerOpts...)
}
