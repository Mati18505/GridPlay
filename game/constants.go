package game

import (
	"math/rand"
)

type char int

const (
	e = iota
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