package game

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGameEnd(t *testing.T) {
	p1 := &Player{}
	p2 := &Player{}
	game := CreateGame(p1, p2)
	game.EndGame = func(winner int) {
		require.Equal(t, winner, p1.GetID())
	}

	require.NotEqual(t, p1.GetChar(), p2.GetChar())
	require.NotEqual(t, p1.GetID(), p2.GetID())

	require.NoError(t, p1.Move(Pos{0,0}))
	require.NoError(t, p2.Move(Pos{1,0}))
	require.NoError(t, p1.Move(Pos{1,1}))
	require.NoError(t, p2.Move(Pos{2,1}))
	require.NoError(t, p1.Move(Pos{2,2}))
	// Game ends here. p1 wins
 	require.Error(t, p2.Move(Pos{1, 2}))
}