import { useState, useCallback } from 'react'
import type { EvalPoint } from '../types/evaluation'

export function useEvaluation() {
  const [evalHistory, setEvalHistory] = useState<EvalPoint[]>([])

  const addEval = useCallback((point: EvalPoint) => {
    setEvalHistory(prev => {
      const filtered = prev.filter(p => p.moveNumber !== point.moveNumber)
      const sorted = [...filtered, point].sort((a, b) => a.moveNumber - b.moveNumber)
      return sorted.length > 200 ? sorted.slice(sorted.length - 200) : sorted
    })
  }, [])

  const reset = useCallback(() => setEvalHistory([]), [])

  return { evalHistory, addEval, reset }
}
