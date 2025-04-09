package chessGame

import (
	"GridPlay/game"
	"testing"

	"github.com/corentings/chess"
	"github.com/stretchr/testify/assert"
)

func TestCreateChessGame(t *testing.T) {
	chessGame := CreateChessGame()

	assert.NotNil(t, chessGame)
	assert.Equal(t, chess.White, chessGame.players[0].color)
	assert.Equal(t, chess.Black, chessGame.players[1].color)
}

func TestValidateMove(t *testing.T) {
	chessGame := CreateChessGame()
	player := chessGame.players[0]

	moveParam := game.MoveParam("e4")
	err := chessGame.ValidateMove(player, moveParam)

	assert.NoError(t, err, "Expected move to be valid")
}

func TestValidateMove_InvalidMove(t *testing.T) {
	chessGame := CreateChessGame()
	player := chessGame.players[0]

	moveParam := game.MoveParam("e9") 
	err := chessGame.ValidateMove(player, moveParam)

	assert.Error(t, err, "Expected move to be invalid")
	assert.Equal(t, "move is not valid", err.Error())
}
func TestValidateMove_EmptyMove(t *testing.T) {
	chessGame := CreateChessGame()
	player := chessGame.players[0]

	moveParam := game.MoveParam("")
	err := chessGame.ValidateMove(player, moveParam)

	assert.Error(t, err, "Expected error for empty move")
	assert.Equal(t, "move is empty", err.Error())
}

func TestMove(t *testing.T) {
	chessGame := CreateChessGame()

	err := chessGame.Move("e4")
	assert.NoError(t, err, "Expected move to succeed")
}

func TestMove_InvalidMove(t *testing.T) {
	chessGame := CreateChessGame()

	err := chessGame.Move("e9")
	assert.Error(t, err, "Expected move to fail")
}

func TestMove_EmptyMove(t *testing.T) {
	chessGame := CreateChessGame()

	err := chessGame.Move("")
	assert.Error(t, err, "Expected error for empty move")
}

func TestGetWinState(t *testing.T) {
	chessGame := CreateChessGame()

	// Simulate a game where white wins
	chessGame.Move("e4")
	chessGame.Move("e5")
	chessGame.Move("Bc4")
	chessGame.Move("Nc6")
	chessGame.Move("Qh5")
	chessGame.Move("Nf6")
	chessGame.Move("Qxf7") // Checkmate

	winState := chessGame.GetWinState()
	assert.Equal(t, game.Win, winState.T)
	assert.Equal(t, chessGame.players[0].id, winState.Data.(game.WinData).Player.GetId())
}

func TestGetWinState_NoWin(t *testing.T) {
	chessGame := CreateChessGame()

	// Simulate a game with no winner
	chessGame.Move("e4")
	chessGame.Move("e5")

	winState := chessGame.GetWinState()
	assert.Equal(t, game.None, winState.T, "Expected no winner")
}

func TestGetWinState_Draw(t *testing.T) {
	chessGame := CreateChessGame()

	// Simulate a draw scenario (e.g., stalemate)
	fenStr := "k1K5/8/8/8/8/8/8/1Q6 w - - 0 1"
	fen, _ := chess.FEN(fenStr)
	game := chess.NewGame(fen)
	game.MoveStr("Qb6")
	game.Method()

	winState := chessGame.GetWinState()
	assert.Equal(t, game.Draw, winState.T, "Expected a draw")
}
