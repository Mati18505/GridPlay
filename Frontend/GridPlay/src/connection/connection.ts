enum ServerMessages {
    GameEnd = 0,
    GameMessageFromServer,
    Approve,
    NotAllowedErr,
}

export interface GameEnd {
    status: string;
    cause: string;
}
export interface GameMsg {
    name: string;
    data: string;
}
export interface Approve {
    approve: boolean;
    reason: string;
}

enum ClientMessages {
    GameMessageToServer = 0,
}

class ServerConnection {
    private socket: WebSocket;
    private status: string;

    onGameEnd?: (this: ServerConnection, ev: GameEnd) => void;
    onGameMsg?: (this: ServerConnection, ev: GameMsg) => void;
    onApprove?: (this: ServerConnection, ev: Approve) => void;
    onNotAllowed?: (this: ServerConnection, ev: string) => void;

    private static instance: ServerConnection;

    private constructor()
    {
        this.socket = new WebSocket("ws://192.168.1.185:4000/ws")

        window.addEventListener("beforeunload", () => {
            this.socket.close();
        });

        this.socket.onerror = this.socketError.bind(this);
        this.socket.onclose = this.socketClose.bind(this);
        this.socket.onmessage = this.socketMessage.bind(this);
        this.status = "Waiting for match...";
    }

    public static get Instance()
    {
        return this.instance || (this.instance = new this());
    }

    socketError(e: Event) {
        this.status = "Cannot connect to server.";
        console.log("Connection error: ", e);
    }

    socketClose(e: CloseEvent) {
        if (e.reason != "") {
            this.status = e.reason;
        } else {
            this.status = "Cannot connect to server.";
        }
        console.log("Connection closed: code: ", e.code, " reason: ", e.reason);
    }

    socketMessage(e: MessageEvent) {
        console.log(e)
        const message = JSON.parse(e.data);
        const messageData = message.data;

        switch (message.type) {
            case ServerMessages.GameEnd:
                const gameEnd: GameEnd = {
                    status: messageData.status,
                    cause: messageData.cause,
                };

                this.onGameEnd?.call(this, gameEnd);
                break;

            case ServerMessages.GameMessageFromServer:
                const gameMsg: GameMsg = {
                    name: messageData.name,
                    data: messageData.data,
                };

                this.onGameMsg?.call(this, gameMsg);
                break;

            case ServerMessages.Approve:
                const approve: Approve = {
                    approve: messageData.approved,
                    reason: messageData.reason,
                };

                this.onApprove?.call(this, approve);
                break;

            case ServerMessages.NotAllowedErr:
                this.onNotAllowed?.call(this, messageData.reason);
                break;

            default:
                console.log("Unknown message type.")
                break;
        }
    }
}

export default ServerConnection;