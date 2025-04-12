package chessGame

import (
	"GridPlay/assert"
	"GridPlay/game"
	"GridPlay/gameServer/externalEvent"
	"bytes"
	"encoding/json"
	"errors"
	"slices"

	"github.com/corentings/chess"
)

type Player struct {
	id int
	color chess.Color
}

func (p Player) GetId() int {
	return p.id
}

type ChessGame struct {
	chess *chess.Game
	players [2]Player
}

func CreateChessGame() *ChessGame {
	players := [2]Player{
		{
			id: 0,
			color: chess.White,
		},
		{
			id: 1,
			color: chess.Black,
		},
	}
	return &ChessGame{
		chess: chess.NewGame(),
		players: players,
	}
}

type msgGameStart struct {
	Color string `json:"color"`
}

func (chessGame *ChessGame) GetGameStartMessage(playerId int) externalEvent.EventGameMessage {
	var data msgGameStart

	if playerId == 0 {
		data = msgGameStart{
			Color: "white",
		}
	} else {
		data = msgGameStart{
			Color: "black",
		}
	}

	return externalEvent.EventGameMessage{Name: "game_start", Data: data}
}

func (chessGame *ChessGame) HandleGameMsg(msg externalEvent.EventGameMessage) ([]externalEvent.EventGameMessage, error) {
	return []externalEvent.EventGameMessage{
		{},
	}, nil
}

func (chessGame *ChessGame) ValidateMove(p Player, m game.MoveParam) error {
	if m == "" {
		return errors.New("move is empty")
	}

	if chessGame.chess.Position().Turn() != p.color {
		return errors.New("not your turn")
	}

	move, err := chessGame.decodeMoveParam(m)
	if err != nil {
		return errors.New("move is not valid")
	}

	validMoves := chessGame.chess.ValidMoves()

	if slices.Contains(validMoves, move) {
		return nil
	}
	return errors.New("move is not valid")
}

func (chessGame *ChessGame) Move(s string) error {
	return chessGame.chess.MoveStr(s)
}

func (chessGame *ChessGame) GetWinState() game.WinState {
	switch chessGame.chess.Outcome() {
	case chess.NoOutcome:
		return game.MakeWinState(game.None, game.NoData{})

	case chess.WhiteWon:
		return game.MakeWinState(game.Win, game.WinData{Player: chessGame.players[0]})

	case chess.BlackWon:
		return game.MakeWinState(game.Win, game.WinData{Player: chessGame.players[1]})

	case chess.Draw:
		return game.MakeWinState(game.Draw, game.NoData{})

	default:
		assert.Never("Invalid chess outcome.")
		return game.MakeWinState(game.None, game.NoData{})
	}
}

func (chessGame *ChessGame) decodeMoveParam(m game.MoveParam) (*chess.Move, error) {
	move, err := chess.AlgebraicNotation{}.Decode(chessGame.chess.Position(), m)
	if err != nil {
		return nil, errors.Join(errors.New("cannot decode move param"), err)
	}

	return move, nil
}

func (chessgame *ChessGame) encodeToJsonString(el any) string {
		var buff bytes.Buffer
		encoder := json.NewEncoder(&buff)

		encoder.Encode(el)
		return buff.String()
}