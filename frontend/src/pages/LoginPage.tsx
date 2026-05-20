import { useEffect } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../hooks/useAuth'
import { LoginForm } from '../components/auth/LoginForm'
import styles from './AuthPage.module.css'

export function LoginPage() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: Location })?.from?.pathname ?? '/'

  useEffect(() => {
    if (user) navigate(from, { replace: true })
  }, [user, navigate, from])

  return (
    <main className={styles.page}>
      <div className={styles.card}>
        <div className={styles.header}>
          <h1 className={styles.title}>♛ Chess</h1>
          <p className={styles.subtitle}>{t('auth.loginSubtitle')}</p>
        </div>
        <LoginForm onSuccess={() => navigate(from, { replace: true })} />
        <p className={styles.footer}>
          {t('auth.noAccount')}{' '}
          <Link to="/register" className={styles.footerLink}>
            {t('auth.register')}
          </Link>
        </p>
      </div>
    </main>
  )
}
