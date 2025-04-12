import './Chess.css'
import { useEffect, useRef, useState } from "react";
import { Chessboard, SparePiece } from "react-chessboard";
import { Piece } from 'react-chessboard/dist/chessboard/types';
import ChessGame from './ChessGame';
import ServerConnection from '../connection/connection';
import { GameMsg } from '../connection/connection';

function Chess() {
  const [state, setState] = useState("start")
  const [orientation, setOrientation] = useState<"white" | "black">("white");
  const serverConnection = useRef(ServerConnection.Instance)

  useEffect(() => {
      serverConnection.current.onGameMsg = function (this: ServerConnection, ev: GameMsg) {
        console.log(ev.name);
        switch (ev.name) {
          case "game_start":
            setOrientation(ev.data === "white" ? "white" : "black")
            break;
        }
      }

    return () => {
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
    <Chessboard position={state} onPieceDrop={onDrop} boardOrientation={orientation}></Chessboard>
  )
}

export default Chess;
