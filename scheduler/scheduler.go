package scheduler

import (
	"context"
	"time"

	"pkg.friendsofgo.tech/neural/middleware"
)

type Class struct {
	taskList taskCtrl
}

func New() Class {
	return Class{
		taskList: newTaskCtrl(),
	}
}

type Task struct {
	when     *time.Duration
	createOn time.Time
	what     taskAction
	promise  chan error
}

func (c Class) NewTask(class middleware.Class, cmd command) *Task {
	t := &Task{
		createOn: time.Now(),
		what: taskAction{
			handler: class,
			command: cmd,
		},
		promise: make(chan error, 1),
	}

	return t
}

func (c Class) Start(ctx context.Context, t *Task) chan error {
	c.taskList.add(ctx, t)

	return t.promise
}

func (t *Task) SetWhen(duration time.Duration) {
	t.when = &duration
}

type command interface{}

type taskAction struct {
	handler middleware.Class
	command command
}

type taskCtrl struct {
	taskMap map[*Task]context.CancelFunc
}

func newTaskCtrl() taskCtrl {
	return taskCtrl{taskMap: make(map[*Task]context.CancelFunc)}
}

func (c *taskCtrl) delete(t *Task) {
	delete(c.taskMap, t)
}

func (c *taskCtrl) add(ctx context.Context, task *Task) {
	_ctx, cancelFunc := context.WithCancel(ctx)
	c.taskMap[task] = cancelFunc

	c.Start(_ctx, task)
}

func (c *taskCtrl) Start(ctx context.Context, t *Task) {
	go func(ctx context.Context) {
		if t.when == nil {
			Now := time.Duration(0)
			t.when = &Now
		}
		select {
		case <-time.After(*t.when):
			t.promise <- t.what.handler.Call(ctx, t.what.command)
			close(t.promise)
			c.delete(t)
		case <-ctx.Done():
			t.promise <- ctx.Err()
		}
	}(ctx)
}
