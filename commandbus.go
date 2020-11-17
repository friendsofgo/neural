package neural

import (
	"context"
	"time"

	"pkg.friendsofgo.tech/neural/commandhandler"
	"pkg.friendsofgo.tech/neural/maphandler"
	"pkg.friendsofgo.tech/neural/middleware"
	"pkg.friendsofgo.tech/neural/scheduler"
)

type Command interface{}

type CommandBus interface {
	Use(middleware middleware.Handler)
	Dispatch(ctx context.Context, command Command, opts ...DispatchOptions) chan error
}

func WithDelay(duration time.Duration) DispatchOptions {
	return func(t *scheduler.Task) {
		t.SetWhen(duration)
	}
}

type DispatchOptions func(task *scheduler.Task)

type cb struct {
	middlewareList middleware.List
	resolver       maphandler.Resolver
	schedule       scheduler.Class
}

func (c *cb) Use(middleware middleware.Handler) {
	c.middlewareList.Add(middleware)
}

func (c cb) Dispatch(ctx context.Context, command Command, opts ...DispatchOptions) chan error {
	errChan := make(chan error, 1)
	defer close(errChan)

	cHandler, err := c.resolver.Resolve(command)
	if err != nil {
		errChan <- err

		return errChan
	}

	t := c.schedule.NewTask(c.middlewareList.BuildWith(middleware.Wrap(cHandler)), command)

	for _, opt := range opts {
		opt(t)
	}

	return c.schedule.Start(ctx, t)
}

func New(handlers ...commandhandler.Class) CommandBus {
	resolver := maphandler.NewResolver()

	for _, h := range handlers {
		resolver.AddHandler(h)
	}

	return &cb{
		middlewareList: make([]middleware.Handler, 0),
		resolver:       resolver,
		schedule:       scheduler.New(),
	}
}
