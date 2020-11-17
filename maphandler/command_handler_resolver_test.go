package maphandler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"pkg.friendsofgo.tech/neural/commandhandler"
	"pkg.friendsofgo.tech/neural/maphandler"
)

func TestMapHandler(t *testing.T) {
	t.Run("Given a commandHandler and a mapHandler resolver", func(t *testing.T) {
		resolver := maphandler.NewResolver()
		commandHandler := commandhandler.New(
			func(ctx context.Context, a struct{}) error { return nil },
		)
		resolver.AddHandler(
			commandHandler,
		)
		t.Run("When the mapHandler resolves the command", func(t *testing.T) {
			retrievedCommandHandler, err := resolver.Resolve(struct{}{})
			t.Run("Then a commandHandler exist", func(t *testing.T) {
				require.NoError(t, err)
				require.Equal(t, commandHandler, retrievedCommandHandler)
			})
		})
	})
}
