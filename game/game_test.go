package game

import (
	"TicTacToe/game/winState"
	"container/list"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGameEnd(t *testing.T) {
	game := CreateGame()

	require.NoError(t, game.Move(Pos{0,0}))
	require.NoError(t, game.Move(Pos{1,0}))
	require.NoError(t, game.Move(Pos{1,1}))
	require.NoError(t, game.Move(Pos{2,1}))
	require.NoError(t, game.Move(Pos{2,2}))
	// Game ends here. p1 wins
	
	state := game.GetWinState()

	require.Equal(t, state, winState.Values.Win)
	winner := state.GetPlayer()

	require.Equal(t, winner.Id, 0) // 0 = id of player 1

 	require.Error(t, game.Move(Pos{1, 2}))
}

func TestGameDraw(t *testing.T) {
	game := CreateGame()

	require.NoError(t, game.Move(Pos{0,0}))
	require.NoError(t, game.Move(Pos{2,0}))
	require.NoError(t, game.Move(Pos{1,0}))
	require.NoError(t, game.Move(Pos{0,1}))
	require.NoError(t, game.Move(Pos{2,1}))
	require.NoError(t, game.Move(Pos{1,1}))
	require.NoError(t, game.Move(Pos{0,2}))
	require.NoError(t, game.Move(Pos{1,2}))
	// Game ends here. draw

	require.Equal(t, game.GetWinState(), winState.Values.Draw)
 	require.Error(t, game.Move(Pos{1, 2}))
}

func TestGameWinChecker(t *testing.T) {
	// Horizontal.
	for i := range 3 {
		chk := createEmptyState()
		chk[0][i] = x
		chk[1][i] = x
		chk[2][i] = x

		moveHistory := list.New()
		moveHistory.PushFront(move{pos: Pos{2, i}, playerID: 0})

		chkg := Game{
			state: chk,
			moveHistory: *moveHistory, // only last move
		}

		require.Equal(t, chkg.checkWinnerByLastMove(), char(x))
	}
	// Vertical.
	for i := range 3 {
		chk := createEmptyState()
		chk[i][0] = x
		chk[i][1] = x
		chk[i][2] = x

		moveHistory := list.New()
		moveHistory.PushFront(move{pos: Pos{i, 2}, playerID: 0})

		chkg := Game{
			state: chk,
			moveHistory: *moveHistory,
		}

		require.Equal(t, chkg.checkWinnerByLastMove(), char(x))
	}
	//
	{
		chk := createEmptyState()
		chk[0][0] = x
		chk[1][1] = x
		chk[2][2] = x

		moveHistory := list.New()
		moveHistory.PushFront(move{pos: Pos{2, 2}, playerID: 0})

		chkg := Game{
			state: chk,
			moveHistory: *moveHistory,
		}

		require.Equal(t, chkg.checkWinnerByLastMove(), char(x))
	}
	{
		chk := createEmptyState()
		chk[0][2] = x
		chk[1][1] = x
		chk[2][0] = x

		moveHistory := list.New()
		moveHistory.PushFront(move{pos: Pos{2, 0}, playerID: 0})

		chkg := Game{
			state: chk,
			moveHistory: *moveHistory,
		}

		require.Equal(t, chkg.checkWinnerByLastMove(), char(x))
	}
}