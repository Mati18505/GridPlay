package game

import (
	"TicTacToe/assert"
	"TicTacToe/game/winState"
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

type Game struct {
	players [2]Player
	state [][]char
	winState winState.WinState
	moveHistory list.List
}

func CreateGame() *Game {
	p1 := Player{
		char: RandomChar(),
		id: 0,
	}
	p2 := Player{
		char: OpponentChar(p1.char),
		id: 1,
	}

	game := &Game{
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

func (game *Game) checkWinnerByLastMove() char {
    var winner char

	state := game.state
	lastMove, err := game.getLastMove()

	if err != nil  {
		return 0
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

    return winner
}

func (game *Game) checkDraw() bool {
    return game.moveHistory.Len() == int(math.Pow(3.0, 2.0)) - 1
}


func (game *Game) Move(pos Pos) error {
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

func (game *Game) check(pos Pos, c char) error {
	if game.state[pos.X][pos.Y] != e {
		return errors.New("cell is not empty")
	}

	game.state[pos.X][pos.Y] = c;

	return nil
}


func (game *Game) GetWinState() winState.WinState {
	return game.winState
}

type move struct {
	pos Pos
	playerID int
}

func (game *Game) GetCurrentRoundPlayer() Player {
	lastMove, err := game.getLastMove()

	if err != nil {
		return game.players[0]
	}

	return game.players[1 - lastMove.playerID]
}

func (game *Game) getLastMove() (move, error) {
	lastMove := game.moveHistory.Front()
	
	if lastMove == nil {
		return move{}, errors.New("there are no moves")
	}

	return lastMove.Value.(move), nil
}

func (game *Game) GetPlayerWithId(id int) Player {
	if id < 0 || id > 1 {
		assert.Never("player id must be 0 or 1", "player id", id)
	}
	return game.players[id]
}