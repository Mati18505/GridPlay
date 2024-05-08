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

func TestGameDraw(t *testing.T) {
	p1 := &Player{}
	p2 := &Player{}
	game := CreateGame(p1, p2)
	game.EndGame = func(winner int) {
		require.Equal(t, winner, -1)
	}

	require.NotEqual(t, p1.GetChar(), p2.GetChar())
	require.NotEqual(t, p1.GetID(), p2.GetID())

	require.NoError(t, p1.Move(Pos{0,0}))
	require.NoError(t, p2.Move(Pos{2,0}))
	require.NoError(t, p1.Move(Pos{1,0}))
	require.NoError(t, p2.Move(Pos{0,1}))
	require.NoError(t, p1.Move(Pos{2,1}))
	require.NoError(t, p2.Move(Pos{1,1}))
	require.NoError(t, p1.Move(Pos{0,2}))
	require.NoError(t, p2.Move(Pos{1,2}))
	// Game ends here. draw
 	require.Error(t, p2.Move(Pos{1, 2}))
}

func TestGameSecurity(t *testing.T) {
	p1 := &Player{}
	p2 := &Player{}

	require.Error(t, p1.Move(Pos{1,1}))

	game := CreateGame(p1, p2)
	game.EndGame = func(winner int) {
		require.Equal(t, winner, p1.GetID())
	}

	require.NotEqual(t, p1.GetChar(), p2.GetChar())
	require.NotEqual(t, p1.GetID(), p2.GetID())

	require.Error(t, p2.Move(Pos{1,0}))
	require.NoError(t, p1.Move(Pos{0,0}))
	require.Error(t, p1.Move(Pos{2,2}))
	require.NoError(t, p2.Move(Pos{2,1}))
	require.NoError(t, p1.Move(Pos{1,1}))
}

func TestGameWinChecker(t *testing.T) {
	// Horizontal.
	for i := range 3 {
		chk := createEmptyState()
		chk[0][i] = x
		chk[1][i] = x
		chk[2][i] = x

		chkg := Game{
			state: chk,
		}

		require.Equal(t, chkg.checkWinnerByLastMove(Pos{1, i}), char(x))
	}
	// Vertical.
	for i := range 3 {
		chk := createEmptyState()
		chk[i][0] = x
		chk[i][1] = x
		chk[i][2] = x

		chkg := Game{
			state: chk,
		}

		require.Equal(t, chkg.checkWinnerByLastMove(Pos{i, 1}), char(x))
	}
	//
	{
		chk := createEmptyState()
		chk[0][0] = x
		chk[1][1] = x
		chk[2][2] = x

		chkg := Game{
			state: chk,
		}

		require.Equal(t, chkg.checkWinnerByLastMove(Pos{1, 1}), char(x))
	}
	{
		chk := createEmptyState()
		chk[0][2] = x
		chk[1][1] = x
		chk[2][0] = x

		chkg := Game{
			state: chk,
		}

		require.Equal(t, chkg.checkWinnerByLastMove(Pos{1, 1}), char(x))
	}
}