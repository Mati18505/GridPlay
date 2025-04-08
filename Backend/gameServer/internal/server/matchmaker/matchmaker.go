package matchmaker

import (
	"TicTacToe/assert"
	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/internal/server"
	"TicTacToe/gameServer/internal/server/serverEvents"

	"github.com/google/uuid"
)

type Matchmaker struct {
	mediator server.Mediator
	matcher chan uuid.UUID
	isLoopRunning bool
	stopLoop chan bool
}

func CreateMatchMaker(mediator server.Mediator) *Matchmaker {
	assert.NotNil(mediator, "mediator was nil")

	return &Matchmaker{
		mediator: mediator,
		matcher: make(chan uuid.UUID, 2),
	}
}

func (mmaker *Matchmaker) StartLoop() {
	assert.Assert(!mmaker.isLoopRunning, "loop was already running")

	go mmaker.loop()
	mmaker.isLoopRunning = true
}

func (mmaker *Matchmaker) EndLoop() {
	assert.Assert(mmaker.isLoopRunning, "loop wasn't running")

	mmaker.stopLoop <- true
	mmaker.isLoopRunning = false
}

func (mmaker *Matchmaker) Add(uuid uuid.UUID) {
	mmaker.matcher <- uuid
}

func (mmaker *Matchmaker) loop() {
	ids := make([]uuid.UUID, 0, 2)

	for {
		select {
		case id := <-mmaker.matcher:
			assert.Assert(len(ids) < 2, "wrong ids length")
			ids = append(ids, id)

			if len(ids) == 2 {
				mmaker.match(ids)
				ids = nil
			} 
		case <-mmaker.stopLoop:
			return
		}
	}
}

func (mmaker *Matchmaker) match(ids []uuid.UUID) {
	assert.Assert(len(ids) == 2, "wrong ids length")

	mmaker.notifyMediator(
		EventPlayersMatched{
			[2]uuid.UUID{ids[0], ids[1]},
		},
	)
}

func (mmaker *Matchmaker) notifyMediator(e event.Event) {
	assert.NotNil(mmaker.mediator, "mediator was nil")

	mmaker.mediator.Notify(serverEvents.MediatorEvent{
		Sender: serverEvents.Matchmaker,
		Event: e,
	})
}