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

// From server
const MatchStarted = 0
const MoveAns = 1
const OpponentMove = 2
const WinEvent = 3

// From client
const Move = 0

var lastMovePos;
var char;
var opponentChar;

// Events
const eventMove = new Event("move");

socket.onmessage = function(messageJson) {
    const message = JSON.parse(messageJson.data);
    const messageData = message.data;

    console.log(message.type);
    switch (message.type) {
        case MoveAns:
            if(messageData.approved) {
                console.log("Move approved");
                document.dispatchEvent(eventMove);
            } else {
                console.log(`Move rejected reason: ${messageData.reason}`)
            }
            
            break;
        case MatchStarted:
            console.log("Match started");
            const eventMatchStarted = new CustomEvent("eventMatchStarted", {
                detail: {
                    char: String.fromCharCode(messageData.char),
                    opponentChar: String.fromCharCode(messageData.opponentChar),
                }
            });
            document.dispatchEvent(eventMatchStarted);
            
            break;
        case WinEvent:
            console.log("Game end: ", messageData.status, messageData.cause);

            const eventWin = new CustomEvent("eventWin", {
                detail: {
                    status: messageData.status,
                }
            });
            document.dispatchEvent(eventWin);
            break;
        case OpponentMove:
            console.log("Opponent move: ", messageData.x, messageData.y);

            const eventOpponentMove = new CustomEvent("opponentMove", {
                detail: {
                    pos: new Pos(messageData.x, messageData.y),
                }
            });
            document.dispatchEvent(eventOpponentMove);
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