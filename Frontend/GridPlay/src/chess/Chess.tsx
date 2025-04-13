import './Chess.css'
import { useEffect, useRef, useState } from "react";
import { Chessboard } from "react-chessboard";
import ServerConnection from '../connection/connection';
import { GameMsg } from '../connection/connection';

function Chess() {
  const [state, setState] = useState("start")
  const [orientation, setOrientation] = useState<"white" | "black">("white");
  const serverConnection = useRef(ServerConnection.Instance)

  useEffect(() => {
    console.log("test")
      serverConnection.current.onGameMsg = function (this: ServerConnection, ev: GameMsg) {
        console.log(ev.name);
        console.log(ev.data);

        switch (ev.name) {
          case "game_start":
            setOrientation(ev.data.color === "white" ? "white" : "black")
            serverConnection.current.sendGameMsg({"foo": "bar"})
            break;
          case "fen_update":
            setState(ev.data.fen)
            break;
        }
      }

    return () => {
      serverConnection.current.onGameMsg = undefined;
    }
  }, [])

  // on opponent move
  function opponentMove() {
    // update fen
  }

  function move(sourceSquare: string, targetSquare: string) {
    sourceSquare = targetSquare
    // chess.move -> approve, fen
    // update fen

    // return approve
  }
  
  function onDrop(sourceSquare: string, targetSquare: string) {
    // move(sourceSquare, targetSquare) -> approve
    // if not approve -> return false

    return true
  }

  return (
    <div className="board_wrapper">
      <Chessboard position={state} onPieceDrop={onDrop} boardOrientation={orientation}></Chessboard>
    </div>
  )
}

export default Chess;
