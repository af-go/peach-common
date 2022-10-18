package http

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/af-go/peach-common/pkg/log"
)

func TestServer(t *testing.T) {
	serverOptions := ServerOptions{Port: 9090, Host: ""}
	logger := log.NewLogger(true)
	hhandler := NewDummyHealthCheckHandler()
	fhandler := NewSimpleFSHandler("./testdata", "/ui")
	Server := NewServer(serverOptions, logger, hhandler, fhandler)
	ctx := context.Background()
	Server.Start(ctx)

	time.Sleep(5 * time.Second)

	clientOptions := ClientOptions{Timeout: 15}
	client := NewClient(clientOptions, logger)
	var response StatusResponse
	headers := make(map[string]string)
	err := client.Get("http://localhost:9090/healthz", headers, &response)
	if err != nil {
		t.Fatalf("failed execute GET request %v", err)
	}
	if response.Message != "Up" {
		t.Fatalf("failed to eval response, expect 'Up', actual: %s", response.Message)
	}
	var result string
	result, err = client.GetRaw("http://localhost:9090/ui", headers)
	if err != nil {
		t.Fatalf("failed execute GET request %v", err)
	}
	if strings.Index("result", "Hello World") > 0 {
		t.Fatalf("failed to eval response, expect 'Hello World', actual: %s", result)
	}
	//Server.Stop(ctx)

}

func TestNoServer(t *testing.T) {
	logger := log.NewLogger(true)
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
}
