package connection

import (
	"log/slog"

	"TicTacToe/assert"
	"TicTacToe/gameServer/message"

	"github.com/gorilla/websocket"
)

type Connection struct {
	socket  *websocket.Conn
	messageFromClient chan message.Message
	exitChan chan bool
	receives bool
}

func CreateConnection(socket *websocket.Conn) *Connection {
	assert.NotNil(socket, "websocket was nil")

	return &Connection{
		socket: socket,
		messageFromClient: make(chan message.Message),
		exitChan: make(chan bool),
		receives: false,
	}
}

func (conn *Connection) StartReceiving() {
	assert.Assert(!conn.receives, "connection was already receiving")

	go conn.receiveMessages()
	conn.receives = true
}

func (conn *Connection) StopReceiving() {
	assert.Assert(conn.receives, "connection wasn't receiving")
	assert.NotNil(conn.socket, "websocket was nil")

	closeMess := websocket.FormatCloseMessage(1000, "Connection closed by server.")
	conn.socket.WriteMessage(websocket.CloseMessage, closeMess)
}

func (conn *Connection) receiveMessages() {
	assert.NotNil(conn.socket, "websocket was nil")

	for {
		_, data, err := conn.socket.ReadMessage()
		if err != nil {
			slog.Info("connection closed with", "ip", conn.GetRemoteIP())
			break;
		}

		msg, err := message.UnmarshalMessage(data)
		if err != nil {
			slog.Warn("cannot unmarshal message, received from", "ip", conn.GetRemoteIP())
			continue
		}

		conn.messageFromClient <- msg
	}

	conn.socket.Close()
	conn.exitChan <- true
	conn.receives = false
}

func (conn *Connection) SendMessage(msg message.Message) error {
	assert.NotNil(msg, "msg was nil")
	assert.NotNil(conn.socket, "websocket was nil")

	data := msg.MarshallMessage()
	err := conn.socket.WriteMessage(websocket.TextMessage, data)

	return err
}

func (conn *Connection) SendPing() error {
	assert.NotNil(conn.socket, "websocket was nil")

	return conn.socket.WriteMessage(websocket.PingMessage, []byte("ping"))
}

func (conn *Connection) GetRemoteIP() string {
	assert.NotNil(conn.socket, "websocket was nil")

	return conn.socket.NetConn().RemoteAddr().String();
}

func (conn *Connection) GetMessageFromClient() <-chan message.Message {
	return conn.messageFromClient
}

func (conn *Connection) GetExitChan() <-chan bool {
	return conn.exitChan
}