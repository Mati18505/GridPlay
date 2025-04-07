package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
func TestMarshalMessage(t *testing.T) {
	msg := MakeMessage(TMatchStarted, MatchStarted{Char: 'X', OpponentChar: 'O'})
	bytes, err := msg.MarshalMessage()
	assert.NoError(t, err)

	var unmarshalled Message
	err = json.Unmarshal(bytes, &unmarshalled)
	assert.NoError(t, err)
	assert.Equal(t, msg.Type, unmarshalled.Type)
	assert.Equal(t, msg.Data, unmarshalled.Data)
}*/

func TestMakeMessage(t *testing.T) {
	data := MatchStarted{Char: 'X', OpponentChar: 'O'}
	msg := MakeMessage(TMatchStarted, data)

	assert.Equal(t, TMatchStarted, msg.Type)
	assert.Equal(t, data, msg.Data)
}

func TestMsgTypeString(t *testing.T) {
	tests := []struct {
		msgType MsgType
		expected string
	}{
		{TMatchStarted, "match_started"},
		{TMoveAns, "move_answer"},
		{TOpponentMove, "opponent_move"},
		{TWinEvent, "win_event"},
		{TNotAllowedErr, "not_allowed_error"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.msgType.String())
	}
}
