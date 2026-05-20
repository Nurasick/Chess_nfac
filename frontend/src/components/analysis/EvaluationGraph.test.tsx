import React from 'react'
import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { EvaluationGraph } from './EvaluationGraph'
import type { EvalPoint } from '../../types/evaluation'

let capturedData: unknown[] = []

vi.mock('recharts', async () => {
  const actual = await vi.importActual<typeof import('recharts')>('recharts')
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    LineChart: ({ data, children }: { data: unknown[]; children: React.ReactNode }) => {
      capturedData = data
      return <div data-testid="line-chart">{children}</div>
    },
  }
})

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      if (opts) return `${key}:${JSON.stringify(opts)}`
      return key
    },
    i18n: { language: 'en', changeLanguage: vi.fn() },
  }),
}))

function makePoint(moveNumber: number, score: number, mate?: number | null): EvalPoint {
  return { moveNumber, score, depth: 10, mate: mate ?? null }
}

describe('EvaluationGraph', () => {
  beforeEach(() => {
    capturedData = []
  })

  it('renders empty state when evalHistory is empty', () => {
    render(<EvaluationGraph evalHistory={[]} />)
    expect(screen.getByText('analysis.noData')).toBeInTheDocument()
  })

  it('renders a LineChart when evalHistory has entries', () => {
    render(<EvaluationGraph evalHistory={[makePoint(1, 0.5)]} />)
    expect(screen.getByTestId('line-chart')).toBeInTheDocument()
  })

  it('data passed to LineChart has shape { move, score, label }', () => {
    render(<EvaluationGraph evalHistory={[makePoint(1, 0.5)]} />)
    expect(capturedData[0]).toMatchObject({ move: 1, score: 0.5, label: '0.50' })
  })

  it('score is clamped to [-10, 10] for high values', () => {
    render(<EvaluationGraph evalHistory={[makePoint(1, 15)]} />)
    expect((capturedData[0] as { score: number }).score).toBe(10)
  })

  it('mate label renders as M2 format', () => {
    render(<EvaluationGraph evalHistory={[makePoint(1, 99, 2)]} />)
    expect((capturedData[0] as { label: string }).label).toBe('M2')
  })

  it('non-mate label renders as score.toFixed(2)', () => {
    render(<EvaluationGraph evalHistory={[makePoint(1, 1.5)]} />)
    expect((capturedData[0] as { label: string }).label).toBe('1.50')
  })
})
