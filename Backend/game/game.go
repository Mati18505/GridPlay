package game

type MoveParam = string

type Game interface {
	ValidateMove(Player, MoveParam) error
	Move(MoveParam) error
	GetWinState() WinState 
}

type Player interface {
	GetId() int
}