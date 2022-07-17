package http

import (
	"context"
	"testing"
	"time"

	"github.com/af-go/peach-common/pkg/log"
)

func TestServer(t *testing.T) {
	serverOptions := ServerOptions{Port: 9090}
	logger := log.NewLogger(true)
	handler := DummyHealthCheckHandler{}
	Server := NewServer(serverOptions, logger, &handler)
	ctx := context.Background()
	Server.Start(ctx)

	time.Sleep(5 * time.Second)

	clientOptions := ClientOptions{Timeout: 15}
	client := NewClient(clientOptions, logger)
	var response StatusResponse
	err := client.Get("http://localhost:9090/healthz", make(map[string]string), &response)
	if err != nil {
		t.Fatalf("failed execute GET request %v", err)
	}
	if response.Message != "Up" {
		t.Fatalf("failed to eval response, expect 'Up', actual: %s", response.Message)
	}
	Server.Stop(ctx)

}
