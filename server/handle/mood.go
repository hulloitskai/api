package handle

import (
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	hr "github.com/julienschmidt/httprouter"
	"github.com/stevenxie/api"
	ess "github.com/unixpickle/essentials"
)

// A MoodsHandler handles requests for mood data.
type MoodsHandler struct {
	svc api.MoodService
	l   *zap.SugaredLogger
}

// NewMoodsHandler creates a new MoodsHandler.
func NewMoodsHandler(svc api.MoodService,
	logger *zap.SugaredLogger) *MoodsHandler {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	return &MoodsHandler{svc: svc, l: logger}
}

const moodsLimitMax = 50

// ListMoods lists moods.
func (mh *MoodsHandler) ListMoods(w http.ResponseWriter, r *http.Request,
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
	moods, err := mh.svc.ListMoods(limit, offset)
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

// GetMood gets a mood with a particular id.
func (mh *MoodsHandler) GetMood(w http.ResponseWriter, r *http.Request,
	ps hr.Params) {
	var (
		id = ps.ByName("id")
		rw = responseWriter{w, mh.l}
	)

	// Get mood by ID.
	mood, err := mh.svc.GetMood(id)
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
