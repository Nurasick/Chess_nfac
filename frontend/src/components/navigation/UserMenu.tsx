import { useState, useRef, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../../hooks/useAuth'
import styles from './UserMenu.module.css'

export function UserMenu() {
  const { t } = useTranslation()
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [isOpen, setIsOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  if (!user) return null

  return (
    <div className={styles.container} ref={ref}>
      <button
        className={styles.trigger}
        onClick={() => setIsOpen(o => !o)}
        aria-expanded={isOpen}
        aria-haspopup="menu"
      >
        <span className={styles.avatar}>{user.username[0].toUpperCase()}</span>
        <span className={styles.username}>{user.username}</span>
        <span className={styles.rating}>{user.rating}</span>
      </button>
      {isOpen && (
        <div className={styles.dropdown} role="menu">
          <Link to="/profile" className={styles.item} onClick={() => setIsOpen(false)}>
            {t('nav.profile')}
          </Link>
          <Link to="/history" className={styles.item} onClick={() => setIsOpen(false)}>
            {t('nav.history')}
          </Link>
          <hr className={styles.divider} />
          <button
            className={styles.item}
            onClick={async () => { await logout(); setIsOpen(false); navigate('/login') }}
            role="menuitem"
          >
            {t('auth.logout')}
          </button>
        </div>
      )}
    </div>
  )
}
