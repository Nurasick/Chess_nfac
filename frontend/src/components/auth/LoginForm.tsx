import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../../hooks/useAuth'
import { loginSchema, type LoginInput } from '../../utils/validators'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import styles from './AuthForm.module.css'

interface LoginFormProps {
  onSuccess?: () => void
}

export function LoginForm({ onSuccess }: LoginFormProps) {
  const { t } = useTranslation()
  const { login } = useAuth()
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginInput>({ resolver: zodResolver(loginSchema) })

  const onSubmit = async (data: LoginInput) => {
    try {
      await login(data.username, data.password)
      onSuccess?.()
    } catch {
      setError('root', { message: t('errors.invalidCredentials') })
    }
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit(onSubmit)} noValidate>
      <Input
        label={t('auth.username')}
        type="text"
        autoComplete="username"
        error={errors.username?.message}
        {...register('username')}
      />
      <Input
        label={t('auth.password')}
        type="password"
        autoComplete="current-password"
        error={errors.password?.message}
        {...register('password')}
      />
      {errors.root && <p className={styles.rootError}>{errors.root.message}</p>}
      <Button type="submit" fullWidth loading={isSubmitting}>
        {t('auth.login')}
      </Button>
    </form>
  )
}
