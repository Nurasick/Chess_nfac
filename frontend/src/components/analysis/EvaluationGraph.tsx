import { useTranslation } from 'react-i18next'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ReferenceLine,
  ResponsiveContainer,
} from 'recharts'
import type { EvalPoint } from '../../types/evaluation'
import styles from './EvaluationGraph.module.css'

interface EvaluationGraphProps {
  evalHistory: EvalPoint[]
}

export function EvaluationGraph({ evalHistory }: EvaluationGraphProps) {
  const { t } = useTranslation()

  if (evalHistory.length === 0) {
    return (
      <div className={styles.empty}>
        <span>{t('analysis.noData')}</span>
      </div>
    )
  }

  const data = evalHistory.map(p => ({
    move: p.moveNumber,
    score: Math.max(-10, Math.min(10, p.score)),
    label: p.mate != null ? `M${p.mate}` : p.score.toFixed(2),
  }))

  return (
    <div className={styles.container} data-testid="eval-graph">
      <ResponsiveContainer width="100%" height={120}>
        <LineChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: -20 }}>
          <XAxis
            dataKey="move"
            tick={{ fontSize: 10, fill: 'var(--color-text-tertiary)' }}
            tickLine={false}
            axisLine={{ stroke: 'var(--color-border-subtle)' }}
          />
          <YAxis
            domain={[-10, 10]}
            tick={{ fontSize: 10, fill: 'var(--color-text-tertiary)' }}
            tickLine={false}
            axisLine={false}
            tickCount={5}
          />
          <Tooltip
            contentStyle={{
              background: 'var(--color-bg-elevated)',
              border: '1px solid var(--color-border-default)',
              borderRadius: '6px',
              fontSize: '12px',
              color: 'var(--color-text-primary)',
            }}
            formatter={(value: unknown, _: unknown, props: { payload?: { label?: string } }) =>
              [props.payload?.label ?? String(value), t('analysis.eval')]
            }
            labelFormatter={(label) => `${t('analysis.move')} ${label}`}
          />
          <ReferenceLine y={0} stroke="var(--color-border-default)" strokeDasharray="4 2" />
          <Line
            type="monotone"
            dataKey="score"
            stroke="var(--color-primary)"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4, fill: 'var(--color-primary-light)' }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
