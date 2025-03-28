package connection

import (
	"log"

	"TicTacToe/gameServer/message"

	"github.com/gorilla/websocket"
)

type Connection struct {
	socket  *websocket.Conn
	messageFromClient chan *message.Message
	exitChan chan bool
}

func CreateConnection(socket *websocket.Conn) *Connection {
	return &Connection{
		socket: socket,
		messageFromClient: make(chan *message.Message),
		exitChan: make(chan bool),
	}
}

// Blocking
func (conn *Connection) ReceiveMessages() {
	for {
		_, data, err := conn.socket.ReadMessage()
		if err != nil {
			log.Printf("connection with %q closed\n", conn.GetRemoteIP())
			conn.exitChan <- true
			break;
		}

		msg, err := message.UnmarshalMessage(data)
		if err != nil {
			log.Println("cannot unmarshal message received from ", conn.GetRemoteIP())
			continue
		}

		conn.messageFromClient <- msg
	}
}

func (conn *Connection) SendMessage(msg *message.Message) error {
	data := msg.MarshallMessage()

	err := conn.socket.WriteMessage(websocket.TextMessage, data)
	return err
}

func (conn *Connection) SendPing() error {
	return conn.socket.WriteMessage(websocket.PingMessage, []byte("ping"))
}

func (conn *Connection) Close() {
	// TODO: Wait until messages are send?
	conn.socket.Close()
}

func (conn *Connection) GetRemoteIP() string {
	return conn.socket.NetConn().RemoteAddr().String();
}

func (conn *Connection) GetMessageFromClient() <-chan *message.Message {
	return conn.messageFromClient
}

func (conn *Connection) GetExitChan() <-chan bool {
	return conn.exitChan
}