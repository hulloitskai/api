package handle

import (
	"net/http"
	"os"

	"go.uber.org/zap"

	httprouter "github.com/julienschmidt/httprouter"
	"github.com/stevenxie/api/internal/info"
)

type apiInfo struct {
	Name    string  `json:"name"`
	Version string  `json:"version"`
	Env     *string `json:"environment"`
}

// An InfoHandler handles requests for API server information.
type InfoHandler struct {
	template *apiInfo
	l        *zap.SugaredLogger
}

// NewInfoHandler creates a new InfoHandler.
func NewInfoHandler(logger *zap.SugaredLogger) *InfoHandler {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	return &InfoHandler{
		template: &apiInfo{
			Name:    info.Namespace,
			Version: info.Version,
		},
		l: logger,
	}
}

func (ih *InfoHandler) cloneTemplate() *apiInfo {
	info := *ih.template
	return &info
}

func (ih *InfoHandler) info() *apiInfo {
	info := ih.cloneTemplate()
	if env, ok := os.LookupEnv("GOENV"); ok {
		info.Env = &env
	}
	return info
}

// GetInfo is an httprouter.Handle that responds with information about the
// API server.
func (ih *InfoHandler) GetInfo(w http.ResponseWriter, _ *http.Request,
	_ httprouter.Params) {
	rw := responseWriter{w, ih.l}
	_ = rw.WriteJSON(ih.info())
}
