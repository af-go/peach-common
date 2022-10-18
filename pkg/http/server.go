package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

type Handler interface {
	Build(engine *gin.Engine)
}

// Options http server options
type ServerOptions struct {
	Host            string `json:"host" yaml:"host"`
	Port            int    `json:"port" yaml:"port"`
	CAFile          string `json:"caFile" yaml:"caFile"`
	PrivateKeyFile  string `json:"privateKetFile" yaml:"privateKeyFile"`
	PublicCertFile  string `json:"publicCertFile" yaml:"publicCertFile"`
	EnableProfiling bool   `json:"enableProfiling" yaml:"enableProfiling"`
}

// NewServer create new Server
func NewServer(options ServerOptions, logger *logr.Logger, handlers ...Handler) *Server {
	return &Server{options: options, logger: logger, handlers: handlers}
}

// Server http Server
type Server struct {
	options  ServerOptions
	logger   *logr.Logger
	server   *http.Server
	handlers []Handler
}

// Start start http Server
func (c *Server) Start(ctx context.Context) {
	gin.SetMode(gin.ReleaseMode)
	port := c.options.Port
	if port == 0 {
		port = 8080
	}
	r := gin.Default()
	for _, h := range c.handlers {
		h.Build(r)
	}
	if c.options.EnableProfiling {
		pprof.Register(r)
	}
	c.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", c.options.Host, port),
		Handler: r,
	}
	go func() {
		if err := c.server.ListenAndServe(); err != http.ErrServerClosed {
			c.logger.Error(err, "failed to start Server", "err", err)
		}
	}()
	c.logger.Info(fmt.Sprintf("Server is listening on port %d", port))
}

// Stop stop API Server
func (c *Server) Stop(ctx context.Context) {
	c.logger.Info(fmt.Sprintf("shutting down Server at %v", time.Now()))
	_, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer func() {
		cancel()
	}()
	if err := c.server.Shutdown(ctx); err != nil {
		c.logger.Error(err, "failed to shut down Server gracefully")
	}
	c.logger.Info(fmt.Sprintf("HTTP Server is shutdown at %v", time.Now()))
}

func NewError(gc *gin.Context, status int, err error) {
	er := HTTPError{
		Code:    status,
		Message: err.Error(),
	}
	gc.JSON(status, er)
}

// HTTPError error response
type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

// StatusResponse status response
type StatusResponse struct {
	Message string `json:"message"`
}
