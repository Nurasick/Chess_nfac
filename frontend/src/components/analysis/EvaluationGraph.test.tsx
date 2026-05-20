import React from 'react'
import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { EvaluationGraph } from './EvaluationGraph'
import type { EvalPoint } from '../../types/evaluation'

let capturedData: unknown[] = []
let capturedFormatter: ((v: unknown, _: unknown, p: { payload?: { label?: string } }) => unknown) | undefined
let capturedLabelFormatter: ((label: unknown) => unknown) | undefined

vi.mock('recharts', async () => {
  const actual = await vi.importActual<typeof import('recharts')>('recharts')
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    LineChart: ({ data, children }: { data: unknown[]; children: React.ReactNode }) => {
      capturedData = data
      return <div data-testid="line-chart">{children}</div>
    },
    Tooltip: ({ formatter, labelFormatter }: {
      formatter?: (v: unknown, _: unknown, p: { payload?: { label?: string } }) => unknown
      labelFormatter?: (label: unknown) => unknown
    }) => {
      capturedFormatter = formatter
      capturedLabelFormatter = labelFormatter
      return null
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
    capturedFormatter = undefined
    capturedLabelFormatter = undefined
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

  it('Tooltip formatter returns payload label when available', () => {
    render(<EvaluationGraph evalHistory={[makePoint(1, 0.5)]} />)
    const result = capturedFormatter?.(0.5, undefined, { payload: { label: '0.50' } })
    expect(result).toEqual(['0.50', 'analysis.eval'])
  })

  it('Tooltip formatter falls back to String(value) when no payload label', () => {
    render(<EvaluationGraph evalHistory={[makePoint(1, 0.5)]} />)
    const result = capturedFormatter?.(0.5, undefined, {})
    expect(result).toEqual(['0.5', 'analysis.eval'])
  })

  it('Tooltip labelFormatter returns move label', () => {
    render(<EvaluationGraph evalHistory={[makePoint(1, 0.5)]} />)
    const result = capturedLabelFormatter?.(3)
    expect(result).toBe('analysis.move 3')
  })
})
