package gameServer

import (
	"log"

	"github.com/gorilla/websocket"
)

type Connection struct {
	socket  *websocket.Conn
	messageFromClient chan *message
	exitChan chan bool
}

func CreateConnection(socket  *websocket.Conn) *Connection {
	return &Connection{
		socket: socket,
		messageFromClient: make(chan *message),
		exitChan: make(chan bool),
	}
}

// Blocking
func (conn *Connection) receiveMessages() {
	for {
		_, data, err := conn.socket.ReadMessage()
		if err != nil {
			log.Printf("connection with %q closed\n", conn.GetRemoteIP())
			close(conn.exitChan)
			break;
		}

		msg, err := UnmarshalMessage(data)
		if err != nil {
			log.Println("cannot unmarshal message received from ", conn.GetRemoteIP())
			continue
		}

		conn.messageFromClient <- msg
	}
}

func (conn *Connection) sendMessage(msg *message) error {
	data, err := MarshallMessage(msg)
	if err != nil {
		return err
	}

	err = conn.socket.WriteMessage(websocket.TextMessage, data)
	return err
}

func (conn *Connection) sendPing() error {
	return conn.socket.WriteMessage(websocket.PingMessage, []byte("ping"))
}

func (conn *Connection) close() {
	// TODO: Wait until messages are send?
	conn.socket.Close()
}

func (conn *Connection) GetRemoteIP() string {
	return conn.socket.NetConn().RemoteAddr().String();
}