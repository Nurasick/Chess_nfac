import { useRef, useCallback, useEffect, useState } from 'react'
import type { EvalPoint } from '../types/evaluation'
import { STOCKFISH_DEBOUNCE_MS } from '../utils/constants'

interface StockfishState {
  isReady: boolean
  currentEval: EvalPoint | null
  bestMove: string | null
}

export function useStockfish() {
  const workerRef = useRef<Worker | null>(null)
  const cacheRef = useRef(new Map<string, EvalPoint>())
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const evalHandlersRef = useRef<Array<(e: EvalPoint) => void>>([])
  const [state, setState] = useState<StockfishState>({ isReady: false, currentEval: null, bestMove: null })

  useEffect(() => {
    let worker: Worker
    try {
      worker = new Worker('/stockfish.js')
      workerRef.current = worker

      worker.onmessage = (e: MessageEvent<string>) => {
        const line = e.data
        if (line === 'uciok') {
          worker.postMessage('isready')
        } else if (line === 'readyok') {
          setState(prev => ({ ...prev, isReady: true }))
        } else if (line.startsWith('bestmove')) {
          const parts = line.split(' ')
          if (parts[1] && parts[1] !== '(none)') {
            setState(prev => ({ ...prev, bestMove: parts[1] }))
          }
        } else if (line.startsWith('info depth')) {
          const depthMatch = line.match(/depth (\d+)/)
          const cpMatch = line.match(/score cp (-?\d+)/)
          const mateMatch = line.match(/score mate (-?\d+)/)
          if (depthMatch) {
            const depth = parseInt(depthMatch[1])
            let score = 0
            let mate: number | null = null
            if (cpMatch) score = parseInt(cpMatch[1]) / 100
            if (mateMatch) { mate = parseInt(mateMatch[1]); score = mate > 0 ? 99 : -99 }
            const evalPoint: EvalPoint = { moveNumber: 0, score, depth, mate }
            setState(prev => ({ ...prev, currentEval: evalPoint }))
            evalHandlersRef.current.forEach(h => h(evalPoint))
          }
        }
      }

      worker.postMessage('uci')
    } catch {
      // Stockfish not available (no WASM in env)
    }

    return () => {
      workerRef.current?.terminate()
      workerRef.current = null
    }
  }, [])

  const analyze = useCallback((fen: string, depth = 20) => {
    const cached = cacheRef.current.get(fen)
    if (cached) {
      setState(prev => ({ ...prev, currentEval: cached }))
      evalHandlersRef.current.forEach(h => h(cached))
      return
    }

    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      if (!workerRef.current || !state.isReady) return
      workerRef.current.postMessage(`position fen ${fen}`)
      workerRef.current.postMessage(`go depth ${depth}`)
    }, STOCKFISH_DEBOUNCE_MS)
  }, [state.isReady])

  const onEval = useCallback((handler: (e: EvalPoint) => void) => {
    evalHandlersRef.current.push(handler)
    return () => {
      evalHandlersRef.current = evalHandlersRef.current.filter(h => h !== handler)
    }
  }, [])

  const stop = useCallback(() => {
    workerRef.current?.postMessage('stop')
  }, [])

  return { ...state, analyze, onEval, stop }
}
