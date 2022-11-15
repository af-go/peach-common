package pool

import (
	"fmt"
	"testing"
	"time"

	"github.com/af-go/peach-common/pkg/log"
	"github.com/go-logr/logr"
)

var names = []string{
	"steve",
	"bob",
	"mary",
	"therese",
	"jason",
}

type PrinterTask struct {
	data   string
	logger *logr.Logger
}

// Work implements the Worker interface.
func (m *PrinterTask) Run(id int) (interface{}, error) {
	m.logger.Info(fmt.Sprintf("Printing %v", m.data))
	time.Sleep(time.Second)
	return nil, nil
}

func (m *PrinterTask) GetName() string {
	return "PrinterTask"
}

func TestPool(t *testing.T) {
	logger := log.NewLogger(true)
	pool := New(20, 20, 1*time.Second, logger)

	//wg := sync.WaitGroup{}
	//wg.Add(10 * len(names))
	for i := 0; i < 10; i++ {
		for _, name := range names {
			printer := PrinterTask{data: name, logger: logger}
			//go func() {
			logger.Info("dispatching task", "task", printer.GetName())
			pool.Dispatch(&printer)
			logger.Info("task is dispatched", "task", printer.GetName())
			//}()
			//wg.Done()
		}
	}
	//wg.Wait()
	pool.Shutdown()
	result := *pool.GetResult()
	if len(result) != 10*len(names) {
		t.Errorf("Failed, Except %v , actual %v ", 10*len(names), len(result))
	}
}
