package TicTacToe

import (
	"GridPlay/assert"
	"GridPlay/game/winState"
	"container/list"
	"errors"
	"math"
)

type Player struct {
	char char
	id int
}

func (p *Player) GetID() int {
	return p.id
}

func (p *Player) GetChar() char {
	return p.char
}

type Pos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type TicTacToe struct {
	players [2]Player
	state [][]char
	winState winState.WinState
	moveHistory list.List
}

func CreateGame() *TicTacToe {
	p1 := Player{
		char: RandomChar(),
		id: 0,
	}
	p2 := Player{
		char: OpponentChar(p1.char),
		id: 1,
	}

	game := &TicTacToe{
		players: [2]Player{p1, p2},
		state: createEmptyState(),
		winState: winState.Values.None,
		moveHistory: *list.New(),
	}

	return game
}

func createEmptyState() [][]char {
	rows := make([][]char, 3)

	for i := range rows {
		rows[i] = make([]char, 3)
	}

	return rows
}

func areEqual[t comparable](v1, v2, v3 t) bool {
	return v1 == v2 && v2 == v3;
}

func (game *TicTacToe) checkWinnerByLastMove() char {
    var winner char

	state := game.state
	lastMove, err := game.getLastMove()

	if err != nil  {
		return e
	}

	pos := lastMove.pos

    if areEqual(state[pos.X][0], state[pos.X][1], state[pos.X][2]) && state[pos.X][0] != e {
        winner = state[pos.X][0]
    }
    if areEqual(state[0][pos.Y], state[1][pos.Y], state[2][pos.Y]) && state[0][pos.Y] != e {
        winner = state[0][pos.Y]
    }
    if areEqual(state[0][0], state[1][1], state[2][2]) && state[0][0] != e {
        winner = state[0][0]
    }
    if areEqual(state[2][0], state[1][1], state[0][2]) && state[2][0] != e {
        winner = state[2][0]
    }

	assert.Assert(winner >= 0 && winner <= 2, "winner out of range", "winner", winner)
    return winner
}

func (game *TicTacToe) checkDraw() bool {
    return game.moveHistory.Len() == int(math.Pow(3.0, 2.0))
}

func (game *TicTacToe) Move(pos Pos) error {
	if pos.X < 0 || pos.Y < 0 || pos.X > 2 || pos.Y > 2 {
		assert.Never("position is out of range", "pos", pos)
	}

	if game.winState != winState.Values.None {
		return errors.New("cannot move after game ended")
	}
		
	p := game.GetCurrentRoundPlayer()
	err := game.check(pos, p.char)

	if err != nil {
		return err
	}

	game.moveHistory.PushFront(move{
		pos: pos,
		playerID: p.id,
	})

	if game.checkWinnerByLastMove() != e {
		state := winState.Values.Win
		state.Player = winState.Player{Id: p.id, Char: int(p.char)}

		game.winState = state
	} else if game.checkDraw() {
		game.winState = winState.Values.Draw
	}

	return nil
}

func (game *TicTacToe) check(pos Pos, c char) error {
	if pos.X < 0 || pos.Y < 0 || pos.X > 2 || pos.Y > 2 {
		assert.Never("position is out of range", "pos", pos)
	}
	assert.Assert(c >= 0 && c <= 2, "char out of range", "char", c)

	if game.state[pos.X][pos.Y] != e {
		return errors.New("cell is not empty")
	}

	game.state[pos.X][pos.Y] = c;

	return nil
}


func (game *TicTacToe) GetWinState() winState.WinState {
	return game.winState
}

type move struct {
	pos Pos
	playerID int
}

func (game *TicTacToe) GetCurrentRoundPlayer() Player {
	lastMove, err := game.getLastMove()

	if err != nil {
		return game.players[0]
	}

	return game.players[1 - lastMove.playerID]
}

func (game *TicTacToe) getLastMove() (move, error) {
	lastMove := game.moveHistory.Front()
	
	if lastMove == nil {
		return move{}, errors.New("there are no moves")
	}

	m, ok := lastMove.Value.(move)
	assert.Assert(ok, "type assertion failed for value move")

	return m, nil
}

func (game *TicTacToe) GetPlayerWithId(id int) Player {
	if id < 0 || id > 1 {
		assert.Never("player id must be 0 or 1", "player id", id)
	}
	return game.players[id]
}