package jobserver

import (
	"github.com/gocraft/work"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/processing"
)

func (srv *Server) registerMoodFetcher() {
	fetcher := processing.NewMoodFetcher(srv.provider, srv.provider)
	fetcher.SetLogger(srv.l.Named("moodfetcher"))
	srv.moodFetcher = fetcher

	pool := work.NewWorkerPool(empty{}, 1, info.Namespace, srv.redisPool)
	pool.Job("fetch_moods", moodFetcher{fetcher}.FetchMoods)

	// TODO: finish this up, make sure your default timings are set in the
	// function itself dont default in constructor
	pool.PeriodicallyEnqueue(srv.fetchMoodsCron, "fetch_moods")
	srv.workerPools["moodfetcher"] = pool
}

type (
	moodFetcher struct{ *processing.MoodFetcher }
	empty       struct{}
)

func (mf moodFetcher) FetchMoods(*work.Job) error {
	return mf.MoodFetcher.FetchMoods()
}
