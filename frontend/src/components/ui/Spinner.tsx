import styles from './Spinner.module.css'

interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  center?: boolean
  label?: string
}

export function Spinner({ size = 'md', center = false, label = 'Loading...' }: SpinnerProps) {
  const spinner = (
    <span
      className={[styles.spinner, styles[size]].join(' ')}
      role="status"
      aria-label={label}
    />
  )

  if (center) {
    return <div className={styles.center}>{spinner}</div>
  }

  return spinner
}
