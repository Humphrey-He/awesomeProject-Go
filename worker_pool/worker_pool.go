package worker_pool

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrPoolClosed indicates the pool no longer accepts new tasks.
	ErrPoolClosed = errors.New("worker pool is closed")
	// ErrPoolStopped indicates the pool was force-stopped.
	ErrPoolStopped = errors.New("worker pool is stopped")
)

// Task defines a unit of work executed by workers.
type Task func(ctx context.Context) (any, error)

// Result carries execution outcome for a submitted task.
type Result struct {
	WorkerID int
	Value    any
	Err      error
}

type request struct {
	ctx    context.Context
	task   Task
	result chan Result
}

// Pool limits concurrency for processing many tasks.
type Pool struct {
	taskCh chan request

	poolCtx context.Context
	cancel  context.CancelFunc

	mu      sync.RWMutex
	closed  bool
	stopNow bool

	once sync.Once
	wg   sync.WaitGroup
}

// New creates a pool with fixed worker count and bounded queue size.
func New(workerCount, queueSize int) *Pool {
	if workerCount <= 0 {
		workerCount = 1
	}
	if queueSize < 0 {
		queueSize = 0
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		taskCh:   make(chan request, queueSize),
		poolCtx:  ctx,
		cancel:   cancel,
		stopNow:  false,
		closed:   false,
		once:     sync.Once{},
		wg:       sync.WaitGroup{},
		mu:       sync.RWMutex{},
	}

	p.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go p.worker(i + 1)
	}
	return p
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for req := range p.taskCh {
		select {
		case <-p.poolCtx.Done():
			req.result <- Result{WorkerID: id, Err: ErrPoolStopped}
			close(req.result)
			continue
		default:
		}

		runCtx, cancel := mergeContext(p.poolCtx, req.ctx)
		value, err := req.task(runCtx)
		cancel()

		req.result <- Result{
			WorkerID: id,
			Value:    value,
			Err:      err,
		}
		close(req.result)
	}
}

// Submit enqueues a task and returns a one-shot result channel.
func (p *Pool) Submit(ctx context.Context, task Task) (<-chan Result, error) {
	if task == nil {
		return nil, errors.New("task must not be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	p.mu.RLock()
	closed := p.closed
	p.mu.RUnlock()
	if closed {
		return nil, ErrPoolClosed
	}

	resultCh := make(chan Result, 1)
	req := request{
		ctx:    ctx,
		task:   task,
		result: resultCh,
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.poolCtx.Done():
		return nil, ErrPoolStopped
	case p.taskCh <- req:
		return resultCh, nil
	}
}

// Shutdown stops accepting tasks and waits queued/running tasks to finish.
func (p *Pool) Shutdown(ctx context.Context) error {
	p.once.Do(func() {
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()
		close(p.taskCh)
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		p.wg.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// Stop immediately cancels running tasks and rejects all future submissions.
func (p *Pool) Stop() {
	p.mu.Lock()
	p.closed = true
	p.stopNow = true
	p.mu.Unlock()
	p.cancel()
	p.once.Do(func() {
		close(p.taskCh)
	})
}

func mergeContext(parentA, parentB context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-parentA.Done():
		case <-parentB.Done():
		case <-ctx.Done():
			return
		}
		cancel()
	}()
	return ctx, cancel
}

