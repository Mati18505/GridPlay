package gameServer

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

type playerConnection struct {
	id uuid.UUID
	connection *Connection
	playerChan chan<- Event
	serverChan chan<- Event
	stopLoop chan bool
}

func CreatePlayerConnection(uuid uuid.UUID, conn *Connection, serverChan chan<- Event) *playerConnection {
	return &playerConnection{
		id: uuid,
		connection: conn,
		playerChan: nil,
		serverChan: serverChan,
		stopLoop: make(chan bool),
	}
}

func (playerConn *playerConnection) StartLoop() {
	go playerConn.loop()
}

func (playerConn *playerConnection) EndLoop() {
	playerConn.stopLoop <- true
}

func (pConn *playerConnection) loop() {
	conn := pConn.connection
	remoteIP := conn.GetRemoteIP()

	LOOP:
	for {
		select {
		case msg := <- conn.messageFromClient:

			log.Printf("playerConnection: received message from %q: Type: %v, ", remoteIP, ClientMsg(msg.Type))

			eType, err := eventTypeFromMessage(msg)

			if err != nil {
				log.Printf("Unknown type of message")
				continue
			} 
			
			e := CreateEvent(eType)

			log.Printf("playerConnection: Created event %+v", e)

			err = pConn.sendEventToPlayerChan(e)
				
			if err != nil {
				message, err := MakeMessage(NotAllowedErr, &NotAllowedErrMessage{
					Reason: "cannot do this while game is not running",
				})

				if err != nil {
					log.Print("cannot make not allowed err message")
					continue;
				}
	
				conn.sendMessage(message)
				log.Printf("cannot do this while game is not running")
			}

		case <- conn.exitChan:
			log.Printf("disconnected from %q\n", remoteIP)
			conn.close()

			e := CreateEvent(EventExit{
				connectionId: pConn.id,
			})
			err := pConn.sendEventToPlayerChan(e)

			if err != nil {
				e := CreateEvent(EventExit{
					connectionId: pConn.id,
				})
				pConn.serverChan <- e
			}

			break LOOP
		case <- pConn.stopLoop:
			break LOOP
		}
	}
}

func (pConn *playerConnection) sendEventToPlayerChan(e Event) error {
	if pConn.playerChan == nil {
		return errors.New("player connection channel does not exist")
	}

	pConn.playerChan <- e
	return nil
}