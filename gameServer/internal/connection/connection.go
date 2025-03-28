package connection

import (
	"log"

	"TicTacToe/assert"
	"TicTacToe/gameServer/message"

	"github.com/gorilla/websocket"
)

type Connection struct {
	socket  *websocket.Conn
	messageFromClient chan *message.Message
	exitChan chan bool
	receives bool
}

func CreateConnection(socket *websocket.Conn) *Connection {
	assert.NotNil(socket, "websocket was nil")

	return &Connection{
		socket: socket,
		messageFromClient: make(chan *message.Message),
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

	conn.socket.Close()
}

func (conn *Connection) receiveMessages() {
	assert.NotNil(conn.socket, "websocket was nil")

	for {
		_, data, err := conn.socket.ReadMessage()
		if err != nil {
			log.Printf("connection with %q closed\n", conn.GetRemoteIP())
			break;
		}

		msg, err := message.UnmarshalMessage(data)
		if err != nil {
			log.Println("cannot unmarshal message received from ", conn.GetRemoteIP())
			continue
		}

		conn.messageFromClient <- msg
	}

	conn.exitChan <- true
	conn.receives = false
}

func (conn *Connection) SendMessage(msg *message.Message) error {
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

func (conn *Connection) GetMessageFromClient() <-chan *message.Message {
	return conn.messageFromClient
}

func (conn *Connection) GetExitChan() <-chan bool {
	return conn.exitChan
}