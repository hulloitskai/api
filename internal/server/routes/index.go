package routes

import (
	"net/http"
	"os"

	"go.uber.org/zap"

	hr "github.com/julienschmidt/httprouter"
	"github.com/stevenxie/api/internal/info"
)

func registerIndex(r *hr.Router, logger *zap.SugaredLogger) {
	ih := newIndexHandler(logger)
	ih.RegisterTo(r)
}

type apiInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"environment"`
}

type indexHandler struct {
	template apiInfo
	l        *zap.SugaredLogger
}

func newIndexHandler(logger *zap.SugaredLogger) *indexHandler {
	ai := apiInfo{
		Name:    info.Namespace,
		Version: info.Version,
	}
	return &indexHandler{
		template: ai,
	}
}

func (ih *indexHandler) Info() *apiInfo {
	info := ih.template
	info.Env = os.Getenv("GOENV")
	return &info
}

func (ih *indexHandler) RegisterTo(r *hr.Router) {
	r.GET("/", ih.Handle)
	r.HEAD("/", ih.Handle)
}

func (ih *indexHandler) Handle(w http.ResponseWriter, _ *http.Request,
	_ hr.Params) {
	rw := responseWriter{w, ih.l}
	rw.WriteJSON(ih.Info())
}
