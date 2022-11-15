package probe

import (
	"testing"

	"github.com/af-go/peach-common/pkg/log"
	model "github.com/af-go/peach-common/pkg/model/database"
	"github.com/af-go/peach-common/pkg/utils"
)

func TestClient(t *testing.T) {
	var options model.MySQLOptions
	logger := log.NewLogger(true)
	err := utils.Load("testdata/config.json", &options, logger)
	if err != nil {
		t.Fatalf("failed to load config file %v", err)
	}
	client := BuildMySQLProbe(options, logger)
	if client == nil {
		t.Fatalf("failed to build client %v", err)
	}
	result := client.Do()
	if !result {
		t.Fatalf("mysql is unhealth")
	}
}
