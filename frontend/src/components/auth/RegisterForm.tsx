import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../../hooks/useAuth'
import { registerSchema, type RegisterInput } from '../../utils/validators'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { CITIES } from '../../utils/constants'
import styles from './AuthForm.module.css'

interface RegisterFormProps {
  onSuccess?: () => void
}

export function RegisterForm({ onSuccess }: RegisterFormProps) {
  const { t } = useTranslation()
  const { register: registerUser } = useAuth()
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RegisterInput>({ resolver: zodResolver(registerSchema) })

  const onSubmit = async (data: RegisterInput) => {
    try {
      await registerUser(data.username, data.email, data.password, data.city)
      onSuccess?.()
    } catch {
      setError('root', { message: t('errors.registrationFailed') })
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
        label={t('auth.email')}
        type="email"
        autoComplete="email"
        error={errors.email?.message}
        {...register('email')}
      />
      <Input
        label={t('auth.password')}
        type="password"
        autoComplete="new-password"
        error={errors.password?.message}
        {...register('password')}
      />
      <div className={styles.field}>
        <label className={styles.label} htmlFor="city">{t('auth.city')}</label>
        <select id="city" className={styles.select} {...register('city')}>
          {CITIES.map(city => (
            <option key={city} value={city}>{t(`cities.${city}`)}</option>
          ))}
        </select>
        {errors.city && <span className={styles.fieldError}>{errors.city.message}</span>}
      </div>
      {errors.root && <p className={styles.rootError}>{errors.root.message}</p>}
      <Button type="submit" fullWidth loading={isSubmitting}>
        {t('auth.register')}
      </Button>
    </form>
  )
}
