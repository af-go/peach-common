package http

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/af-go/peach-common/pkg/log"
)

const (
	port = 9099
)

func TestServer(t *testing.T) {
	serverOptions := ServerOptions{Port: port, Host: ""}
	logger := log.NewLogger(true)
	hhandler := NewDummyHealthyHandler()
	fhandler := NewSimpleFSHandler("./testdata", "/ui")
	Server := NewServer(serverOptions, logger, hhandler, fhandler)
	ctx := context.Background()
	Server.Start(ctx)

	time.Sleep(5 * time.Second)

	clientOptions := ClientOptions{Timeout: 15}
	client := NewClient(clientOptions, logger)
	var response StatusResponse
	headers := make(map[string]string)
	url := fmt.Sprintf("http://localhost:%d/healthz", port)
	err := client.Get(url, headers, &response)
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
