import { useTranslation } from 'react-i18next'
import { TIME_CONTROLS } from '../../utils/constants'
import styles from './TimeControlSelect.module.css'

interface TimeControlSelectProps {
  selected: string
  onChange: (value: string) => void
}

export function TimeControlSelect({ selected, onChange }: TimeControlSelectProps) {
  const { t } = useTranslation()

  return (
    <div className={styles.container}>
      <h3 className={styles.title}>{t('queue.selectTimeControl')}</h3>
      <div className={styles.grid}>
        {TIME_CONTROLS.map(tc => (
          <button
            key={tc.value}
            className={[styles.option, selected === tc.value ? styles.selected : ''].join(' ')}
            onClick={() => onChange(tc.value)}
            type="button"
          >
            <span className={styles.name}>{tc.label}</span>
          </button>
        ))}
      </div>
    </div>
  )
}
