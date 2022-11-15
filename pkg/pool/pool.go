package pool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"gonum.org/v1/gonum/stat"
)

const (
	increaseSignal  = 1
	decreaseSingal  = 2
	shutodownSignal = 3
)

type internalTask struct {
	id         string
	AssignTime time.Time
	StartTime  time.Time
	task       Task
}

// Task task interface
type Task interface {
	GetName() string
	Run(workRoutine int) (interface{}, error)
}

// TaskStatus task status
type TaskStatus struct {
	Name       string
	ID         string
	Err        error
	Data       interface{}
	Executor   string
	AssignTime time.Time
	StartTime  time.Time
	Duration   time.Duration
}

// Pool work pool
type Pool struct {
	logger          *logr.Logger
	minRoutines     int
	maxRoutines     int
	controlChannel  chan int
	shutdownChannel chan struct{}
	killChannel     chan struct{} //Channel used to kill goroutines
	taskChannel     chan internalTask
	statusChannel   chan TaskStatus
	routines        int64 //number of current totla available routines
	sequence        int
	wg              sync.WaitGroup
	tasksStatus     []TaskStatus
	count           int64 //number to total tasks
	timeout         time.Duration
	shuttingDown    bool
	taskSequence    int64
}

// New create new pool
func New(minRoutines int, maxRoutines int, timeout time.Duration, logger *logr.Logger) *Pool {
	pool := Pool{
		minRoutines:     minRoutines,
		maxRoutines:     maxRoutines,
		controlChannel:  make(chan int),
		taskChannel:     make(chan internalTask),
		shutdownChannel: make(chan struct{}),
		killChannel:     make(chan struct{}),
		statusChannel:   make(chan TaskStatus),
		tasksStatus:     []TaskStatus{},
		timeout:         timeout,
		logger:          logger,
	}
	pool.init()
	return &pool
}

func (p *Pool) init() {
	if p.minRoutines < 1 {
		p.minRoutines = 1
	}
	if p.maxRoutines < p.minRoutines {
		p.minRoutines = p.maxRoutines
	}
	p.startDaemon()
	p.add(p.minRoutines)
	p.resultRoutine()
}

func (p *Pool) resultRoutine() {
	p.wg.Add(1)
	go func() {
		for {
			select {
			case s := <-p.statusChannel:
				p.tasksStatus = append(p.tasksStatus, s)
				p.logger.V(0).Info("task is done", "task", s.Name)
				atomic.AddInt64(&p.count, -1)
			case <-p.killChannel:
				p.logger.Info("summay channel is killed")
				p.wg.Done()
			}
		}
	}()
}

// add create new routines
func (p *Pool) add(routines int) {
	p.logger.Info("tried to add routines", "routines", routines)
	if routines <= 0 {
		p.logger.V(0).Info("negative routines, ignore", "routines", routines)
		return
	}
	for i := 0; i < routines; i++ {
		p.controlChannel <- increaseSignal
	}
}

// Dispatch submit a task
func (p *Pool) Dispatch(task Task) string {
	if p.shuttingDown {
		p.logger.Info("Shutting down, no request accept")
		return ""
	}
	s := int(atomic.LoadInt64(&p.taskSequence))
	atomic.AddInt64(&p.taskSequence, 1)
	t := internalTask{id: fmt.Sprintf("%v", s), task: task, AssignTime: time.Now()}
	p.taskChannel <- t
	p.logger.Info("task assigned", "task", task.GetName())
	atomic.AddInt64(&p.count, 1)
	return t.id
}

func (p *Pool) execute(id int) {
	p.logger.Info("start routine", "routine", id)
	for {
		select {
		case t := <-p.taskChannel:
			p.logger.Info("run task in routine", "task", t.id, "routine", id)
			t.StartTime = time.Now()
			result := TaskStatus{ID: t.id, Name: t.task.GetName(), AssignTime: t.AssignTime}
			result.Data, result.Err = t.task.Run(id)
			result.Executor = fmt.Sprintf("routine %v", id)
			result.Duration = time.Since(t.StartTime)
			p.statusChannel <- result
			p.logger.Info("task in routine is done", "task", t.id, "routine", id)
		case <-p.killChannel:
			atomic.AddInt64(&p.routines, -1)
			p.wg.Done()
			p.logger.Info("routine is killed", "routine", id)
			return
		}
	}

}

// startDaemon start daemon routine for pool
func (p *Pool) startDaemon() {
	p.wg.Add(1)
	go func() {
		for {
			select {
			case c := <-p.controlChannel:
				switch c {
				case increaseSignal:
					p.logger.Info("increase signal is received")
					p.sequence++
					p.wg.Add(1)
					atomic.AddInt64(&p.routines, 1)
					go p.execute(p.sequence)

				case decreaseSingal:
					p.logger.Info("decrease signal is received")
				}
			case <-p.shutdownChannel:
				p.logger.Info("shutdown signal is received")
				p.shuttingDown = true
				var c int

				for i := 0; i < 30; i++ {
					c = int(atomic.LoadInt64(&p.count))
					p.logger.Info("running tasks", "tasks", c)
					if c == 0 {
						break
					}
					// use mean value of tasks execution time, in case some tasks are very slowly
					averageDuration, err := time.ParseDuration(fmt.Sprintf("%vs", p.mean()))
					if err != nil {
						time.Sleep(p.timeout)
					} else {
						time.Sleep(averageDuration)
					}
				}
				if c > 0 {
					p.logger.Error(fmt.Errorf("timeout"), "tasks", c)
				}
				routines := int(atomic.LoadInt64(&p.routines))
				p.logger.Info("working routines, killing", "routines", routines)
				for i := 0; i < routines; i++ {
					p.killChannel <- struct{}{}
				}
				p.killChannel <- struct{}{}
				p.wg.Done()
				return
			}
		}
	}()
}

func (p *Pool) mean() float64 {
	s := []float64{}
	for _, taskStatus := range p.tasksStatus {
		s = append(s, taskStatus.Duration.Seconds())
	}
	return stat.Mean(s, nil)
}

// Shutdown stutdown pool
func (p *Pool) Shutdown() {
	p.shutdownChannel <- struct{}{}
	p.wg.Wait()
}

// GetResult get tasks result of pool
func (p *Pool) GetResult() *[]TaskStatus {
	return &p.tasksStatus
}
