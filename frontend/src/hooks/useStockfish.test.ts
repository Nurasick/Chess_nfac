import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { useStockfish } from './useStockfish'

interface MockWorker {
  postMessage: ReturnType<typeof vi.fn>
  terminate: ReturnType<typeof vi.fn>
  onmessage: ((e: MessageEvent<string>) => void) | null
}

let mockWorker: MockWorker

class MockWorkerClass {
  postMessage = vi.fn()
  terminate = vi.fn()
  onmessage: ((e: MessageEvent<string>) => void) | null = null

  constructor() {
    mockWorker = this as unknown as MockWorker
  }
}

function fireMessage(data: string) {
  act(() => {
    mockWorker.onmessage?.({ data } as MessageEvent<string>)
  })
}

function makeReady() {
  fireMessage('uciok')
  fireMessage('readyok')
}

describe('useStockfish', () => {
  beforeEach(() => {
    vi.stubGlobal('Worker', MockWorkerClass)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.useRealTimers()
  })

  it('creates a worker with /stockfish.js on mount', () => {
    renderHook(() => useStockfish())
    expect(mockWorker).toBeDefined()
  })

  it('sends uci on mount', () => {
    renderHook(() => useStockfish())
    expect(mockWorker.postMessage).toHaveBeenCalledWith('uci')
  })

  it('after readyok, analyze(fen) debounces 100ms then sends position fen', () => {
    vi.useFakeTimers()
    const { result } = renderHook(() => useStockfish())
    makeReady()

    act(() => {
      result.current.analyze('rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1')
    })

    expect(mockWorker.postMessage).not.toHaveBeenCalledWith(
      expect.stringContaining('position fen')
    )

    act(() => {
      vi.advanceTimersByTime(100)
    })

    expect(mockWorker.postMessage).toHaveBeenCalledWith(
      'position fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'
    )
  })

  it('after readyok, analyze(fen) sends go depth 20 after debounce', () => {
    vi.useFakeTimers()
    const { result } = renderHook(() => useStockfish())
    makeReady()

    act(() => {
      result.current.analyze('rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1')
    })

    act(() => {
      vi.advanceTimersByTime(100)
    })

    expect(mockWorker.postMessage).toHaveBeenCalledWith('go depth 20')
  })

  it('info score cp message sets currentEval.score to centipawns/100', () => {
    const { result } = renderHook(() => useStockfish())
    makeReady()

    fireMessage('info depth 15 score cp 35')

    expect(result.current.currentEval?.score).toBe(0.35)
  })

  it('info score mate message sets currentEval.mate', () => {
    const { result } = renderHook(() => useStockfish())
    makeReady()

    fireMessage('info depth 10 score mate 2')

    expect(result.current.currentEval?.mate).toBe(2)
  })

  it('onEval handler is called with the parsed EvalPoint', () => {
    const { result } = renderHook(() => useStockfish())
    makeReady()

    const handler = vi.fn()
    act(() => {
      result.current.onEval(handler)
    })

    fireMessage('info depth 12 score cp 50')

    expect(handler).toHaveBeenCalledWith(
      expect.objectContaining({ score: 0.5, depth: 12 })
    )
  })

  it('onEval unsubscribe removes the handler', () => {
    const { result } = renderHook(() => useStockfish())
    makeReady()

    const handler = vi.fn()
    let unsub: (() => void) | undefined
    act(() => {
      unsub = result.current.onEval(handler)
    })
    act(() => {
      unsub?.()
    })

    fireMessage('info depth 12 score cp 50')

    expect(handler).not.toHaveBeenCalled()
  })

  it('calls worker.terminate() on unmount', () => {
    const { unmount } = renderHook(() => useStockfish())
    unmount()
    expect(mockWorker.terminate).toHaveBeenCalled()
  })

  it('stop() sends stop to the worker', () => {
    const { result } = renderHook(() => useStockfish())
    act(() => {
      result.current.stop()
    })
    expect(mockWorker.postMessage).toHaveBeenCalledWith('stop')
  })
})
