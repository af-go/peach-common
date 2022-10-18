package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewDummyHealthCheckHandler() *DummyHealthCheckHandler {
	return &DummyHealthCheckHandler{}
}

// DummyHealthCheckHandler dummy health check handler
type DummyHealthCheckHandler struct {
}

// Build build health check handler
func (h *DummyHealthCheckHandler) Build(engine *gin.Engine) {
	engine.GET("/healthz", h.Healthz)
}

// Healthz health check api
// @Produce json
// @Summary health check
// @Description check status
// @Success 200 {object} StatusResponse
// @Failure 400 {object} TTPError
// @Failure 503 {object} HTTPError
// @Router /healthz [get]
func (h *DummyHealthCheckHandler) Healthz(gc *gin.Context) {
	statusCode := 200
	resp := StatusResponse{Message: "Up"}
	gc.JSON(statusCode, &resp)
}

func NewSimpleFSHandler(fsPath string, relativePath string) *SimpleFSHandler {
	return &SimpleFSHandler{fsPath: fsPath, relativePath: relativePath}
}

type SimpleFSHandler struct {
	relativePath string
	fsPath       string
}

// Build build health check handler
func (f *SimpleFSHandler) Build(engine *gin.Engine) {
	engine.StaticFS(f.relativePath, http.Dir(f.fsPath))
}
