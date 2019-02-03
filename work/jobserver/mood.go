package jobserver

import (
	wk "github.com/gocraft/work"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/work"
)

// RegisterMoodFetcher registers a MoodFetcher to the server.
func (s *Server) RegisterMoodFetcher(mf *work.MoodFetcher) {
	ctx := moodFetcherContext{mf}
	pool := wk.NewWorkerPool(ctx, 1, info.Namespace, s.RedisPool)
	pool.Job("fetch_moods", ctx.FetchMoods)
	pool.PeriodicallyEnqueue(s.Config.FetchMoodsCron, "fetch_moods")
	s.WorkerPools["moodfetcher"] = pool
}

type moodFetcherContext struct{ *work.MoodFetcher }

func (mfc moodFetcherContext) FetchMoods(*wk.Job) error {
	return mfc.MoodFetcher.FetchMoods()
}
