package game

import (
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

func (p *Player) GetChar() rune {
	if p.char == x {
		return 'x'
	} else {
		return 'o'
	}
}

type Pos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Game struct {
	players [2]Player
	state [][]char
	turn int 
	moveCount uint
	gameEnded bool
	EndGame func(winner int)
	moveHistory list.List
}

func CreateGame(player1, player2 *Player) *Game {
	player1.char = RandomChar()
	player2.char = OpponentChar(player1.char)
	player1.id = 0
	player2.id = 1

	game := &Game{
		players: [2]Player{*player1, *player2},
		state: createEmptyState(),
		turn: 0,
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
    return game.moveCount == uint(math.Pow(3.0, 2.0)) - 1
}


func (game *Game) Move(pos Pos) error {
	if game.gameEnded {
		return errors.New("cannot move after game ended")
	}
		
	p := game.GetCurrentRoundPlayer()
	err := game.check(pos, p.char, p.id)

	if err != nil {
		return err
	}

	game.moveHistory.PushFront(move{
		pos: pos,
		playerID: p.id,
	})

	if game.checkWinnerByLastMove() != e {
		game.gameEnded = true
		game.EndGame(p.id)
	} else if game.checkDraw() {
		game.gameEnded = true
		game.EndGame(-1)
	}

	return nil
}

func (game *Game) check(pos Pos, c char, pId int) error {
	if game.state[pos.X][pos.Y] != e {
		return errors.New("cell is not empty")
	}

	game.state[pos.X][pos.Y] = c;

	return nil
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