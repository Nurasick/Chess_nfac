import { Link, NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../../hooks/useAuth'
import { UserMenu } from './UserMenu'
import { Button } from '../ui/Button'
import styles from './Header.module.css'

export function Header() {
  const { t, i18n } = useTranslation()
  const { user } = useAuth()

  const toggleLang = () => {
    i18n.changeLanguage(i18n.language === 'ru' ? 'kk' : 'ru')
  }

  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        <Link to="/" className={styles.logo}>
          ♛ Chess
        </Link>
        <nav className={styles.nav} aria-label="Main navigation">
          <NavLink
            to="/leaderboard"
            className={({ isActive }) => [styles.link, isActive ? styles.active : ''].join(' ')}
          >
            {t('nav.leaderboard')}
          </NavLink>
          {user && (
            <>
              <NavLink
                to="/history"
                className={({ isActive }) => [styles.link, isActive ? styles.active : ''].join(' ')}
              >
                {t('nav.history')}
              </NavLink>
            </>
          )}
        </nav>
        <div className={styles.actions}>
          <button className={styles.langToggle} onClick={toggleLang} title="Switch language">
            {i18n.language === 'ru' ? 'ҚАЗ' : 'РУС'}
          </button>
          {user ? (
            <UserMenu />
          ) : (
            <div className={styles.authLinks}>
              <Link to="/login">
                <Button variant="ghost" size="sm">{t('auth.login')}</Button>
              </Link>
              <Link to="/register">
                <Button size="sm">{t('auth.register')}</Button>
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
