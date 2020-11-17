package neural_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"pkg.friendsofgo.tech/neural"
	"pkg.friendsofgo.tech/neural/commandhandler"
	"pkg.friendsofgo.tech/neural/middleware"
	"pkg.friendsofgo.tech/neural/middleware/multierror"
)

//nolint:funlen
func TestCommandBus(t *testing.T) {
	t.Run("Given a command a handler and a command bus", func(t *testing.T) {
		handler := commandhandler.New(func(ctx context.Context, a struct{}) error {
			return nil
		})
		commandBus := neural.New(handler)
		command := struct{}{}
		t.Run("When command goes into the bus", func(t *testing.T) {
			err := commandBus.Dispatch(context.Background(), command)
			t.Run("Then handler is executed", func(t *testing.T) {
				require.NoError(t, <-err)
			})
		})
	})

	t.Run("Given a command a handler and a command bus and a middleware", func(t *testing.T) {
		handlerErr := fmt.Errorf("handler")       //nolint:goerr113
		middlewareErr := fmt.Errorf("middleware") //nolint:goerr113
		handler := commandhandler.New(func(ctx context.Context, a struct{}) error {
			return handlerErr
		})
		commandBus := neural.New(handler)

		commandBus.Use(
			middleware.HandlerFunc(
				func(ctx context.Context, command middleware.Command, next middleware.NextFn) error {
					return multierror.New(
						next(ctx, command),
						middlewareErr,
					)
				},
			),
		)
		command := struct{}{}
		t.Run("When command goes into the bus", func(t *testing.T) {
			errChan := commandBus.Dispatch(context.Background(), command)
			t.Run("Then handler and middleware are executed", func(t *testing.T) {
				err := <-errChan
				require.True(t, errors.Is(err, middlewareErr))
				require.True(t, errors.Is(err, handlerErr))
			})
		})
	})

	t.Run("Given a command a handler and schedule time of 0.5 secs", func(t *testing.T) {
		const duration = 500 * time.Millisecond

		handler := commandhandler.New(func(ctx context.Context, a struct{}) error {
			return nil
		})
		commandBus := neural.New(handler)
		cmd := struct{}{}

		t.Run("When When command goes into the bus ", func(t *testing.T) {
			startTime := time.Now()
			errChan := commandBus.Dispatch(context.Background(), cmd, neural.WithDelay(duration))
			t.Run("Then handler is executed in 0.5 sec", func(t *testing.T) {
				<-errChan
				sub := -time.Until(startTime)
				require.True(t, sub >= duration, sub)
			})
		})
	})
}
