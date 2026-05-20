import type { HTMLAttributes, ReactNode } from 'react'
import styles from './Card.module.css'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode
  padded?: boolean
  elevated?: boolean
  interactive?: boolean
}

export function Card({
  children,
  padded = true,
  elevated = false,
  interactive = false,
  className = '',
  ...props
}: CardProps) {
  const classes = [
    styles.card,
    padded ? styles.padded : '',
    elevated ? styles.elevated : '',
    interactive ? styles.interactive : '',
    className,
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div className={classes} {...props}>
      {children}
    </div>
  )
}
