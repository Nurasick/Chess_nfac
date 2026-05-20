import { renderHook, act } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { useEvaluation } from './useEvaluation'
import type { EvalPoint } from '../types/evaluation'

function makePoint(moveNumber: number, score = 0): EvalPoint {
  return { moveNumber, score, depth: 10, mate: null }
}

describe('useEvaluation', () => {
  it('initial evalHistory is empty array', () => {
    const { result } = renderHook(() => useEvaluation())
    expect(result.current.evalHistory).toEqual([])
  })

  it('addEval appends an EvalPoint to the history', () => {
    const { result } = renderHook(() => useEvaluation())
    act(() => {
      result.current.addEval(makePoint(1, 0.5))
    })
    expect(result.current.evalHistory).toHaveLength(1)
    expect(result.current.evalHistory[0].moveNumber).toBe(1)
  })

  it('addEval with same moveNumber replaces the existing entry', () => {
    const { result } = renderHook(() => useEvaluation())
    act(() => {
      result.current.addEval(makePoint(1, 0.5))
    })
    act(() => {
      result.current.addEval(makePoint(1, 1.0))
    })
    expect(result.current.evalHistory).toHaveLength(1)
    expect(result.current.evalHistory[0].score).toBe(1.0)
  })

  it('addEval keeps history sorted by moveNumber ascending', () => {
    const { result } = renderHook(() => useEvaluation())
    act(() => {
      result.current.addEval(makePoint(3, 0.3))
    })
    act(() => {
      result.current.addEval(makePoint(1, 0.1))
    })
    act(() => {
      result.current.addEval(makePoint(2, 0.2))
    })
    expect(result.current.evalHistory.map(p => p.moveNumber)).toEqual([1, 2, 3])
  })

  it('evalHistory is capped at 200 entries', () => {
    const { result } = renderHook(() => useEvaluation())
    act(() => {
      for (let i = 1; i <= 201; i++) {
        result.current.addEval(makePoint(i, 0))
      }
    })
    expect(result.current.evalHistory).toHaveLength(200)
    // should keep the latest 200 (moves 2-201, not move 1)
    expect(result.current.evalHistory[0].moveNumber).toBe(2)
  })

  it('reset clears evalHistory to empty array', () => {
    const { result } = renderHook(() => useEvaluation())
    act(() => {
      result.current.addEval(makePoint(1, 0.5))
    })
    act(() => {
      result.current.reset()
    })
    expect(result.current.evalHistory).toEqual([])
  })
})
