import { useContext } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowUpCircleIcon } from 'lucide-react'
import { ToastContext } from '@/contexts/ToastContext'
import { UpdateContext } from '@/contexts/UpdateContext'
import Button from '@/components/ui/button'
import { errorText } from '@/error'

/**
 * Strip shown on every tab once a check found a newer version. Installing it replaces the
 * binary and restarts the service, so the interface goes away for a few seconds.
 */
export default function UpdateBanner() {
  const { t } = useTranslation(['system', 'common'])
  const { status, installing } = useContext(UpdateContext).data
  const { install } = useContext(UpdateContext).actions
  const { showToast } = useContext(ToastContext).actions

  if (!status?.available) {
    return null
  }

  const update = async () => {
    try {
      await install()
      showToast(t('update.restarting'), 'success')
    } catch (error) {
      showToast(errorText(error, t('update.failed')), 'danger')
    }
  }

  return (
    <div className="flex flex-wrap items-center justify-between gap-4 rounded-lg border border-brand-orange/40 bg-brand-orange/10 px-5 py-4">
      <div className="flex items-center gap-3">
        <ArrowUpCircleIcon size={20} className="shrink-0 text-brand-orange" />
        <p className="text-[15px]">{t('update.available', { version: status.latest })}</p>
      </div>
      <Button onClick={update} loading={installing}>
        {t('update.install')}
      </Button>
    </div>
  )
}
