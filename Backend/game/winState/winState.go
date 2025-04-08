package winState

// Duplicate, because of cycle import.
type Player struct {
	Char int
	Id int
}

type WinState interface {
	isWinState()
	GetPlayer() *Player
}

type none struct {}
type draw struct {}
type win struct {
	Player Player
}
func (*none) isWinState() {}
func (*draw) isWinState() {}
func (*win) isWinState() {}

func (*none) GetPlayer() *Player { return nil; }
func (*draw) GetPlayer() *Player { return nil; }
func (m *win) GetPlayer() *Player { return &m.Player; }

var (
	Values = struct {
		None *none
		Draw *draw
		Win *win
	}{
		None: &none{},
		Draw: &draw{},
		Win: &win{},
	}
)