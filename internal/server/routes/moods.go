package routes

import (
	"errors"
	"net/http"
	"strconv"

	hr "github.com/julienschmidt/httprouter"
	"github.com/stevenxie/api/pkg/mood"
	ess "github.com/unixpickle/essentials"
	"go.uber.org/zap"
)

func registerMoods(r *hr.Router, repo mood.Repo, logger *zap.SugaredLogger) {
	mh := &moodsHandler{Repo: repo, l: logger}
	mh.RegisterTo(r)
}

type moodsHandler struct {
	mood.Repo
	l *zap.SugaredLogger
}

func (mh *moodsHandler) RegisterTo(r *hr.Router) {
	r.GET("/moods/", mh.Handle)
}

const (
	moodsLimitMax = 50
)

func (mh *moodsHandler) Handle(w http.ResponseWriter, r *http.Request,
	params hr.Params) {
	var (
		limit   = 10
		startID = ""

		rw  = responseWriter{w, mh.l}
		qp  = r.URL.Query()
		err error
	)

	// Parse and validate query params.
	if l := qp.Get("limit"); l != "" {
		if limit, err = strconv.Atoi(l); err != nil {
			ess.AddCtxTo("routes: parsing 'limit' parameter as int", &err)
		}
		if limit <= 0 {
			err = errors.New("routes: limit must be a positive integer")
		}
		if limit > moodsLimitMax {
			limit = moodsLimitMax
		}
	}
	if err != nil {
		var (
			code = http.StatusBadRequest
			jerr = jsonErrorFrom(err, code)
		)
		w.WriteHeader(code)
		rw.WriteJSON(&jerr)
		return
	}

	if sid := qp.Get("startId"); sid != "" {
		startID = sid
	}

	moods, err := mh.Repo.SelectMoods(limit, startID)
	if err != nil {
		var (
			code = http.StatusInternalServerError
			jerr = jsonErrorFrom(err, code)
		)
		w.WriteHeader(code)
		rw.WriteJSON(&jerr)
		return
	}

	rw.WriteJSON(moods)
}
