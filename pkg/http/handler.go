package http

import (
	"net/http"

	"github.com/af-go/peach-common/pkg/http/probe"
	"github.com/gin-gonic/gin"
)

func NewDummyHealthyHandler() *DummyHealthyHandler {
	return &DummyHealthyHandler{}
}

// DummyHealthyHandler dummy health check handler
type DummyHealthyHandler struct {
}

// Build build health check handler
func (h *DummyHealthyHandler) Build(engine *gin.Engine) {
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
func (h *DummyHealthyHandler) Healthz(gc *gin.Context) {
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

func NewSimpleHealthyHandler() *SimpleHealthyHandler {
	return &SimpleHealthyHandler{}
}

// SimpleHealthyHandler simple health check handler
type SimpleHealthyHandler struct {
}

// Build build health check handler
func (h *SimpleHealthyHandler) Build(engine *gin.Engine) {
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
func (h *SimpleHealthyHandler) Healthz(gc *gin.Context) {
	statusCode := 200
	resp := StatusResponse{Message: "Up"}
	gc.JSON(statusCode, &resp)
}

type ProbeManager struct {
	Probes map[string]*probe.Probe
}
