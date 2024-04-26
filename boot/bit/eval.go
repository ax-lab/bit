package bit

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	GlobalTimeout = 15 * time.Second
	EvalTimeout   = 5 * time.Second
)

type EvalQueue struct {
	sync sync.Mutex

	ctx *Context

	started atomic.Bool
	done    chan struct{}

	tasks []*EvalTask

	canStart    chan struct{}
	evalTimeout chan struct{}
	tasksDone   chan *EvalTask

	running atomic.Int64
	pending atomic.Int64

	waitingByTask map[*EvalTask][]*EvalTask
}

func evalQueueNew(ctx *Context) *EvalQueue {
	queue := &EvalQueue{
		ctx:           ctx,
		done:          make(chan struct{}),
		canStart:      make(chan struct{}),
		evalTimeout:   make(chan struct{}),
		tasksDone:     make(chan *EvalTask, 100),
		waitingByTask: make(map[*EvalTask][]*EvalTask),
	}
	return queue
}

func (queue *EvalQueue) Start() {
	if !queue.started.CompareAndSwap(false, true) {
		panic("Eval queue already running")
	}

	var (
		timeoutFlag  atomic.Bool
		timeoutRaise = func(err error) {
			if timeoutFlag.CompareAndSwap(false, true) {
				queue.ctx.AddError(err)
				close(queue.evalTimeout)
			}
		}
	)

	// watch dog
	go func() {
		select {
		case <-queue.done:
			return
		case <-queue.evalTimeout:
			return
		case <-time.After(GlobalTimeout):
			timeoutRaise(fmt.Errorf("global eval timed out after %s", GlobalTimeout))
		}
	}()

	close(queue.canStart)

	// main run
	go func() {
		defer close(queue.done)

		to := time.NewTimer(EvalTimeout)
		for queue.running.Load() > 0 {
			if !to.Stop() {
				<-to.C
			}
			to.Reset(EvalTimeout)

			isDone := false
			select {
			case task := <-queue.tasksDone:
				queue.running.Add(-1)
				close(task.done)
				if task.hasWaiters.Load() {
					pending := queue.waitingByTask[task]
					delete(queue.waitingByTask, task)
					for _, it := range pending {
						if it.waitingCount.Add(-1) == 0 {
							queue.pending.Add(-1)
						}
					}
				}
			case <-to.C:
				timeoutRaise(fmt.Errorf("eval queue timed out after %s", EvalTimeout))
				isDone = true
			case <-queue.evalTimeout:
				isDone = true
			}

			if isDone {
				break
			}

			if running := queue.running.Load(); running > 0 && running == queue.pending.Load() {
				queue.ctx.AddError(fmt.Errorf("eval queue deadlock"))
				break
			}
		}
	}()
}

func (queue *EvalQueue) Wait() {
	<-queue.done
}

type EvalFn func(task *EvalTask)

type EvalTask struct {
	queue *EvalQueue
	eval  EvalFn

	done chan struct{}

	started atomic.Bool

	waitingCount atomic.Int64
	hasWaiters   atomic.Bool
}

type evalWait struct {
	waiting    *EvalTask
	waitingFor *EvalTask
}

func (queue *EvalQueue) NewTask(eval EvalFn) *EvalTask {
	task := &EvalTask{
		queue: queue,
		eval:  eval,
		done:  make(chan struct{}),
	}

	queue.sync.Lock()
	queue.tasks = append(queue.tasks, task)
	queue.sync.Unlock()
	return task
}

func (task *EvalTask) Queue() {
	if !task.started.CompareAndSwap(false, true) {
		return
	}

	task.queue.running.Add(1)

	go func() {
		defer task.flagDone()
		<-task.queue.canStart
		task.eval(task)
	}()
}

func (task *EvalTask) flagDone() {
	task.queue.tasksDone <- task
}
