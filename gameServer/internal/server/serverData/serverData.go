package serverData

import (
	"TicTacToe/assert"
	"TicTacToe/gameServer/internal/handlers"
	"errors"
	"sync"

	"github.com/google/uuid"
)

type ServerData struct {
	connections map[uuid.UUID]*handlers.PlayerConnection
	rooms       map[uuid.UUID]*handlers.Room
	mut sync.Mutex
}

func CreateServerData() *ServerData {
	 return &ServerData{
		connections: make(map[uuid.UUID]*handlers.PlayerConnection),
		rooms: make(map[uuid.UUID]*handlers.Room),
	 }
}

func (srvData *ServerData) AddRoom(room *handlers.Room) {
	assert.NotNil(room, "room was nil")

	srvData.mut.Lock()
	defer srvData.mut.Unlock()

	srvData.rooms[room.GetUUID()] = room
}

func (srvData *ServerData) RemoveRoom(roomUUID uuid.UUID) {
	srvData.mut.Lock()
	defer srvData.mut.Unlock()

	delete(srvData.rooms, roomUUID)
}

func (srvData *ServerData) ForEachRoom(f func (room *handlers.Room)) {
	srvData.mut.Lock()
	defer srvData.mut.Unlock()

	for _, room := range srvData.rooms {
		f(room)
	}
}

func (srvData *ServerData) AddPlayerConnection(uuid uuid.UUID, pConn *handlers.PlayerConnection) {
	assert.NotNil(pConn, "player connection was nil")

	srvData.mut.Lock()
	defer srvData.mut.Unlock()

	srvData.connections[uuid] = pConn
}

func (srvData *ServerData) RemoveConnection(uuid uuid.UUID) {
	srvData.mut.Lock()
	defer srvData.mut.Unlock()

	delete(srvData.connections, uuid)
}

func (srvData *ServerData) GetConnection(id uuid.UUID) (*handlers.PlayerConnection, error) {
	srvData.mut.Lock()
	defer srvData.mut.Unlock()

	conn, ok := srvData.connections[id]

	if !ok {
		return nil, errors.New("connection does not exist")
	}

	return conn, nil
}