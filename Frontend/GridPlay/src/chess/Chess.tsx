import './Chess.css'
import { useEffect, useReducer, useRef, useState } from "react";
import { Chessboard } from "react-chessboard";
import ServerConnection from '../connection/connection';
import { GameMsg } from '../connection/connection';

interface ChessState {
  position: string;
  orientation: 'white' | 'black';
}

interface GameStartAction {
  type: 'game_start';
  orientation: 'white' | 'black';
}

interface FenUpdateAction {
  type: 'fen_update';
  position: string;
}

type ChessAction = GameStartAction | FenUpdateAction;

function reducer(state: ChessState, action: ChessAction): ChessState {
  console.log(action.type);

  switch (action.type) {
    case 'game_start':
      console.log("Current orientation:", state.orientation);
      console.log("New orientation:", action.orientation);

      ServerConnection.Instance.sendGameMsg({ foo: "bar" });
      return {
        ...state,
        orientation: action.orientation,
      };
    case 'fen_update':
      console.log("Current FEN:", state.position);
      console.log("New FEN:", action.position);

      return {
        ...state,
        position: action.position,
      };
    default:
      return state;
  }
}

function Chess() {
  const [state, dispatch] = useReducer(reducer, { position: 'start', orientation: 'white' })

  useEffect(() => {
    console.log("Setting onGameMsg callback")

    ServerConnection.Instance.onGameMsg = function (this: ServerConnection, ev: GameMsg) {
      switch (ev.name) {
        case 'game_start':
          dispatch({type: ev.name, orientation: ev.data.color})
          break;
        case 'fen_update':
          dispatch({type: ev.name, position: ev.data.fen})
          break;
        default:
          console.log('Unknown chesss action.')
          break;
      }
    }

    return () => {
      console.log("Clearing onGameMsg callback");
      ServerConnection.Instance.onGameMsg = undefined;
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
      <Chessboard position={state.position} onPieceDrop={onDrop} boardOrientation={state.orientation}></Chessboard>
    </div>
  )
}

export default Chess;
