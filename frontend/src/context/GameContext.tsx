import { createContext, useContext, useReducer, useCallback, type ReactNode } from 'react'
import type { Color, GameResult, GameStatus } from '../types/game'
import type { EvalPoint } from '../types/evaluation'

interface GameState {
  gameId: string | null
  fen: string
  myColor: Color | null
  status: GameStatus
  result: GameResult | null
  resultReason: string | null
  whiteRating: number
  blackRating: number
  ratingDelta: number
  moveHistory: string[]
  isMyTurn: boolean
  evalHistory: EvalPoint[]
}

type GameAction =
  | { type: 'GAME_START'; gameId: string; fen: string; myColor: Color; whiteRating: number; blackRating: number }
  | { type: 'MOVE_MADE'; fen: string; notation: string; isMyTurn: boolean }
  | { type: 'GAME_END'; result: GameResult; reason: string; whiteRatingDelta: number; blackRatingDelta: number; myColor: Color }
  | { type: 'EVAL_UPDATE'; eval: EvalPoint }
  | { type: 'RESET' }

const initialState: GameState = {
  gameId: null,
  fen: 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
  myColor: null,
  status: 'waiting',
  result: null,
  resultReason: null,
  whiteRating: 0,
  blackRating: 0,
  ratingDelta: 0,
  moveHistory: [],
  isMyTurn: false,
  evalHistory: [],
}

function reducer(state: GameState, action: GameAction): GameState {
  switch (action.type) {
    case 'GAME_START':
      return {
        ...initialState,
        gameId: action.gameId,
        fen: action.fen,
        myColor: action.myColor,
        status: 'active',
        whiteRating: action.whiteRating,
        blackRating: action.blackRating,
        isMyTurn: action.myColor === 'white',
      }
    case 'MOVE_MADE':
      return {
        ...state,
        fen: action.fen,
        moveHistory: [...state.moveHistory, action.notation],
        isMyTurn: action.isMyTurn,
      }
    case 'GAME_END': {
      const myDelta = action.myColor === 'white' ? action.whiteRatingDelta : action.blackRatingDelta
      return {
        ...state,
        status: 'finished',
        result: action.result,
        resultReason: action.reason,
        ratingDelta: myDelta,
      }
    }
    case 'EVAL_UPDATE':
      return { ...state, evalHistory: [...state.evalHistory, action.eval] }
    case 'RESET':
      return initialState
    default:
      return state
  }
}

interface GameContextValue {
  state: GameState
  dispatch: React.Dispatch<GameAction>
  resetGame: () => void
}

const GameContext = createContext<GameContextValue | null>(null)

export function GameProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initialState)
  const resetGame = useCallback(() => dispatch({ type: 'RESET' }), [])
  return (
    <GameContext.Provider value={{ state, dispatch, resetGame }}>
      {children}
    </GameContext.Provider>
  )
}

export function useGameContext(): GameContextValue {
  const ctx = useContext(GameContext)
  if (!ctx) throw new Error('useGameContext must be used within GameProvider')
  return ctx
}
