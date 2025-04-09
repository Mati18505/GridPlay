package TicTacToe

import (
	"GridPlay/assert"
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
	var opponent char
	
	switch c {
	case x:
		opponent = o
	case o:
		opponent = x
	case e:
		assert.Never("cannot get e opponent")
	default:
		assert.Never("unkown char type", "char", c)
	}

	return opponent
}

func (c char) GetRune() rune {
	var r rune

	switch c {
	case x:
		r = 'x'
	case o:
		r = 'o'
	case e:
		r = ' '
	default:
		assert.Never("unknown char type", "char", c)
	}

	return r
}
