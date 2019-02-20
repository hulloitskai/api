package server

import (
	"github.com/stevenxie/api/server/handle"
)

// registerRoutes registers the routes for srv.router.
func (srv *Server) registerRoutes() {
	r := srv.router

	// Register info handler.
	info := handle.NewInfoHandler(srv.l.Named("info"))
	r.HEAD("/", info.GetInfo)
	r.GET("/", info.GetInfo)

	// Register mood handler.
	moods := handle.NewMoodsHandler(srv.provider, srv.l.Named("moods"))
	r.GET("/moods", moods.ListMoods)
	r.GET("/moods/:id", moods.GetMood)
}
