package game

type WinStateType int 
const (
	None WinStateType = iota
	Draw
	Win
)

type WinState struct {
	T WinStateType
	Data any
}

type NoData struct {}
type WinData struct {
	Player Player
}

func WrapWinState(t WinStateType, data any) WinState {
	return WinState{
		T: t,
		Data: data,
	}
}

func MakeWinState[T any](t WinStateType, data T) WinState {
	msg := WrapWinState(t, data)

	return msg
}