package routes

import (
	"errors"
	"net/http"
	"strconv"

	hr "github.com/julienschmidt/httprouter"
	"github.com/stevenxie/api"
	ess "github.com/unixpickle/essentials"
	"go.uber.org/zap"
)

func registerMoods(r *hr.Router, svc api.MoodService,
	logger *zap.SugaredLogger) {
	mh := &moodsHandler{Svc: svc, l: logger}
	mh.RegisterTo(r)
}

type moodsHandler struct {
	Svc api.MoodService
	l   *zap.SugaredLogger
}

func (mh *moodsHandler) RegisterTo(r *hr.Router) {
	r.GET("/moods/", mh.ListMoods)
	r.GET("/moods/:id", mh.GetMood)
}

const (
	moodsLimitMax = 50
)

func (mh *moodsHandler) ListMoods(w http.ResponseWriter, r *http.Request,
	_ hr.Params) {
	var (
		limit  = 10
		offset int

		rw  = responseWriter{w, mh.l}
		qp  = r.URL.Query()
		err error
	)

	// Parse and validate query params.
	if l := qp.Get("limit"); l != "" {
		if limit, err = strconv.Atoi(l); err != nil {
			ess.AddCtxTo("routes: parsing 'limit' param as int", &err)
		}
		if limit <= 0 {
			err = errors.New("routes: limit must be a positive int")
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

	if os := qp.Get("offset"); os != "" {
		if offset, err = strconv.Atoi(os); err != nil {
			ess.AddCtxTo("routes: parsing 'offset' param as int", &err)
		}
		if offset < 0 {
			err = errors.New("routes: offset must be a non-negative int")
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

	// List moods.
	moods, err := mh.Svc.ListMoods(limit, offset)
	if err != nil {
		var (
			code = http.StatusInternalServerError
			jerr = jsonErrorFrom(err, code)
		)
		w.WriteHeader(code)
		rw.WriteJSON(&jerr)
		return
	}

	// Write response.
	rw.WriteJSON(moods)
}

func (mh *moodsHandler) GetMood(w http.ResponseWriter, r *http.Request,
	params hr.Params) {
	var (
		id = params.ByName("id")
		rw = responseWriter{w, mh.l}
	)

	// Get mood by ID.
	mood, err := mh.Svc.GetMood(id)
	if err != nil {
		var (
			code = http.StatusInternalServerError
			jerr = jsonErrorFrom(err, code)
		)
		w.WriteHeader(code)
		rw.WriteJSON(&jerr)
		return
	}

	// Write response.
	rw.WriteJSON(mood)
}
