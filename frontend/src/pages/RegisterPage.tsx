import { useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../hooks/useAuth'
import { RegisterForm } from '../components/auth/RegisterForm'
import styles from './AuthPage.module.css'

export function RegisterPage() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    if (user) navigate('/', { replace: true })
  }, [user, navigate])

  return (
    <main className={styles.page}>
      <div className={styles.card}>
        <div className={styles.header}>
          <h1 className={styles.title}>♛ Chess</h1>
          <p className={styles.subtitle}>{t('auth.registerSubtitle')}</p>
        </div>
        <RegisterForm onSuccess={() => navigate('/', { replace: true })} />
        <p className={styles.footer}>
          {t('auth.hasAccount')}{' '}
          <Link to="/login" className={styles.footerLink}>
            {t('auth.login')}
          </Link>
        </p>
      </div>
    </main>
  )
}
