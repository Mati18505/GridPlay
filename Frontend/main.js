const socket = new WebSocket("ws://192.168.1.185:4000/ws");
GetStatusEl().innerHTML = "Waiting for match...";

socket.onerror=function(event){
    GetStatusEl().innerHTML = "Cannot connect to server.";
    console.log("Connection error: ", event)
}

socket.onclose = function(event) {
    if (event.reason != "") {
        GetStatusEl().innerHTML = event.reason
    } else {
        GetStatusEl().innerHTML = "Cannot connect to server."
    }
    console.log("Connection closed: code: ", event.code, " reason: ", event.reason)
};

const ServerMessages = {
    GameEnded: 0,
    GameMessageFromServer: 1,
    Approve: 2,
    NotAllowedErr: 3,
};

// From client
const ClientMessages = {
    GameMessageToServer: 0,
}

var lastMovePos;
var char;
var opponentChar;

// Events
const eventApprove = new Event("approve");
const eventNotAllowed = new Event("not-allowed");

socket.onmessage = function(messageJson) {
    const message = JSON.parse(messageJson.data);
    const messageData = message.data;

    console.log(message.type);

    switch (message.type) {
        case ServerMessages.gameEnded:
            console.log("Game end: ", messageData.status, messageData.cause);

            const eventGameEnd = new CustomEvent("game-end", {
                detail: {
                    status: messageData.status,
                    cause: messageData.cause
                }
            });
            document.dispatchEvent(eventGameEnd);
            break;

        case ServerMessages.GameMessageFromServer:
            console.log("Match started");

            const eventGameMsg = new CustomEvent("game-msg", {
                detail: {
                    data: messageData
                }
            });
            document.dispatchEvent(eventGameMsg);
            break;

        case ServerMessages.approved:
            if(messageData.approved) {
                console.log(`Last action approved: ${messageData.reason}`);

                document.dispatchEvent(eventApprove);
            } else {
                console.log(`Move rejected reason: ${messageData.reason}`)
            }
            
            break;

        case ServerMessages.NotAllowedErr:
            console.log("Last action not allowed: " + messageData.reason)

            document.dispatchEvent(eventNotAllowed)
            break;
    }
    /*} catch {
        console.error("cannot parse message from server");
        return;
    }*/
    
};

const container = document.querySelector(".container");
CreateGrid(container);

function Pos(x, y) {
    this.x = x;
    this.y = y;
}

function MoveMessage(pos) {
    this.type = Move;
    this.data = pos;
};

function OnClick(pos) {
    return function (evt) {
        var message = new MoveMessage(pos);
        lastMovePos = pos;
        socket.send(JSON.stringify(message));
    };
}

document.addEventListener("eventMatchStarted", e => {
    console.log("Event match started: char: ", e.detail.char, " opponent char: ", e.detail.opponentChar)
    char = e.detail.char;
    opponentChar = e.detail.opponentChar;
    ClearCells();
    GetStatusEl().innerHTML = "Match started!";
});

document.addEventListener("move", e => {
    GetCell(lastMovePos).innerHTML = char;
    GetStatusEl().innerHTML = "Opponent turn";
});

document.addEventListener("opponentMove", e => {
    console.log("Event opponent move: pos: ", e.detail.pos)
    GetCell(e.detail.pos).innerHTML = opponentChar;
    GetStatusEl().innerHTML = "Your turn";
});

document.addEventListener("eventWin", e => {
    console.log("Event win: status: ", e.detail.status)
    let status;
/*
    if(e.detail.winner === WinStatus.player) {
        status = "Player wins!";
    } else if(e.detail.winner === WinStatus.pc) {
        status = "PC wins!";
    } else {
        status = "Tie!";
    }*/
    status = e.detail.status;
    console.log(status);
    GetStatusEl().innerHTML = status;
    gameEnded = true;
});

document.addEventListener("game-ended", e => {
    console.log("game ended")
    console.log(`status: ${e.detail.status}, cause: ${e.detail.cause}`)
});

document.addEventListener("game-msg", e => {
    console.log("game message")
    console.log(`data: ${e.detail.data}`)
});

document.addEventListener("approve", e => {
    console.log("approve")
});

document.addEventListener("not-allowed", e => {
    console.log("not allowed")
});

function GetStatusEl() {
    return document.querySelector(".status");
}

function CreateID(pos) {
    return `${pos.x}_${pos.y}`;
}

function GetCell(pos) {
    const ID = CreateID(pos);
    return document.getElementById(`${ID}`);
}

function CreateCell(pos) {
    let el = document.createElement("div");
    el.classList += "grid_cell";
    el.id = CreateID(pos);
    el.innerHTML = "";
    el.addEventListener("click", OnClick(pos));
    return el;
}

function CreateGrid(container) {
    for(let x = 0; x < 3; x++) {
        for(let y = 0; y < 3; y++) {
            let pos = new Pos(x, y);
            container.appendChild(CreateCell(pos));
        }
    }   
}

function ClearCells() {
    for(let x = 0; x < 3; x++) {
        for(let y = 0; y < 3; y++) {
            GetCell(new Pos(x, y)).innerHTML = ' ';
        }
    }
}