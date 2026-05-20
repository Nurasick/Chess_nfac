import { useState, useCallback, useRef, useEffect } from 'react'
import { ChessWrapper } from '../lib/chess-wrapper'
import type { Color } from '../types/game'

interface ChessBoardState {
  fen: string
  selectedSquare: string | null
  legalMoves: string[]
  lastMove: { from: string; to: string } | null
  isCheck: boolean
  isGameOver: boolean
}

export function useChessBoard(initialFen?: string) {
  const chessRef = useRef<ChessWrapper | null>(null)
  if (chessRef.current === null) {
    chessRef.current = new ChessWrapper()
    if (initialFen) {
      chessRef.current.loadFen(initialFen)
    }
  }

  const lastLoadedFen = useRef(initialFen ?? '')

  const [boardState, setBoardState] = useState<ChessBoardState>(() => ({
    fen: chessRef.current!.getFen(),
    selectedSquare: null,
    legalMoves: [],
    lastMove: null,
    isCheck: chessRef.current!.isCheck(),
    isGameOver: chessRef.current!.isGameOver(),
  }))

  const syncState = useCallback((extra?: Partial<ChessBoardState>) => {
    setBoardState(prev => ({
      ...prev,
      fen: chessRef.current!.getFen(),
      isCheck: chessRef.current!.isCheck(),
      isGameOver: chessRef.current!.isGameOver(),
      ...extra,
    }))
  }, [])

  useEffect(() => {
    if (initialFen && initialFen !== lastLoadedFen.current) {
      lastLoadedFen.current = initialFen
      chessRef.current!.loadFen(initialFen)
      syncState({ selectedSquare: null, legalMoves: [], lastMove: null })
    }
  }, [initialFen, syncState])

  const selectSquare = useCallback((square: string) => {
    const moves = chessRef.current!.getLegalMovesVerbose(square)
    const targets = (moves as Array<{ to: string }>).map(m => m.to)
    setBoardState(prev => ({
      ...prev,
      selectedSquare: targets.length > 0 ? square : null,
      legalMoves: targets,
    }))
  }, [])

  const makeMove = useCallback((from: string, to: string, promotion?: string): boolean => {
    const result = chessRef.current!.validateMove(from, to, promotion)
    if (!result) return false
    syncState({
      selectedSquare: null,
      legalMoves: [],
      lastMove: { from, to },
    })
    return true
  }, [syncState])

  const loadPosition = useCallback((fen: string) => {
    chessRef.current!.loadFen(fen)
    syncState({ selectedSquare: null, legalMoves: [], lastMove: null })
  }, [syncState])

  const applyServerMove = useCallback((move: string, fen: string) => {
    chessRef.current!.loadFen(fen)
    const from = move.slice(0, 2)
    const to = move.slice(2, 4)
    syncState({ selectedSquare: null, legalMoves: [], lastMove: { from, to } })
  }, [syncState])

  const getTurn = useCallback((): Color => chessRef.current!.getTurn(), [])
  const getHistory = useCallback(() => chessRef.current!.getHistory(), [])
  const reset = useCallback(() => {
    chessRef.current!.reset()
    syncState({ selectedSquare: null, legalMoves: [], lastMove: null })
  }, [syncState])

  return {
    ...boardState,
    selectSquare,
    makeMove,
    loadPosition,
    applyServerMove,
    getTurn,
    getHistory,
    reset,
  }
}
