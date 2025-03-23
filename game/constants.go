package game

import (
	"math/rand"
)

type char rune

const (
	e char = iota
	x
	o
)

func RandomChar() char {
	return char(rand.Intn(2) + 1)
}

func OpponentChar(c char) char {
	switch c {
	case x:
		return o
	case o:
		return x
	default:
		return 0
	}
}

