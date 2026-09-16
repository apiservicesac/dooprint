import { useContext, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { HouseIcon, LinkIcon, PlugZapIcon, RadioTowerIcon, Unlink2Icon } from 'lucide-react'
import { OdooStatus, PairWithOdoo, UnpairFromOdoo } from '@/../wailsjs/go/main/App'
import { main } from '@/../wailsjs/go/models'
import { AppContext } from '@/contexts/AppContext'
import { ToastContext } from '@/contexts/ToastContext'
import { errorText } from '@/error'
import { Badge } from '@/components/ui/badge'
import Button from '@/components/ui/button'
import { Field, RadioCard } from '@/components/ui/field'
import Input from '@/components/ui/input'
import Section from '@/components/ui/section'
import Stat from '@/components/ui/stat'
import StatusDot from '@/components/ui/status_dot'

const POLL_INTERVAL = 5000

const BUS_TONES = {
  connected: 'success',
  disconnected: 'danger',
  connecting: 'warning',
  off: 'muted',
} as const

export default function OdooPanel() {
  const { t } = useTranslation(['odoo', 'common'])
  const appContext = useContext(AppContext)
  const toastContext = useContext(ToastContext)
  const [status, setStatus] = useState<main.OdooStatus | null>(null)

  const refresh = () =>
    OdooStatus()
      .then(setStatus)
      .catch((err) => console.error('Cannot read the Odoo status', err))

  useEffect(() => {
    refresh()
    const id = window.setInterval(refresh, POLL_INTERVAL)
    return () => clearInterval(id)
  }, [])

  async function onUnpair() {
    if (!window.confirm(t('unpair.confirm'))) {
      return
    }
    try {
      setStatus(await UnpairFromOdoo())
      toastContext.actions.showToast(t('unpair.done'), 'success')
    } catch (err) {
      toastContext.actions.showToast(errorText(err, t('unpair.failed')), 'danger')
    }
  }

  if (!status) {
    return (
      <Section icon={<PlugZapIcon size={20} />} title={t('connect.title')}>
        <p className="text-sm text-muted-foreground">{t('common:loading')}</p>
      </Section>
    )
  }

  if (!status.paired) {
    return (
      <PairForm
        defaultName={appContext.data.app?.name ?? ''}
        onPaired={(paired) => {
          setStatus(paired)
          toastContext.actions.showToast(t('connect.paired'), 'success')
        }}
        onError={(message) => toastContext.actions.showToast(message, 'danger')}
      />
    )
  }

  const busTone = BUS_TONES[status.busState as keyof typeof BUS_TONES] ?? 'muted'
  const lastPoll =
    status.lastPoll && !status.lastPoll.startsWith('0001')
      ? new Date(status.lastPoll).toLocaleTimeString()
      : '—'

  return (
    <>
      <Section
        icon={<PlugZapIcon size={20} />}
        title={status.boxName || t('common:device')}
        description={status.odooUrl}
        actions={<StatusDot tone={busTone} label={t(`bus.${status.busState}`, { defaultValue: status.busState })} />}
      >
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          <Stat
            label={t('status.mode')}
            value={
              <Badge variant={status.mode === 'agent' ? 'default' : 'muted'}>
                {status.mode === 'agent' ? <RadioTowerIcon size={12} /> : <HouseIcon size={12} />}
                {status.mode === 'agent' ? t('mode.agent') : t('mode.local')}
              </Badge>
            }
          />
          <Stat label={t('status.printed')} value={status.printed} />
          <Stat label={t('status.failed')} value={status.failed} />
          <Stat label={t('status.lastSeen')} value={lastPoll} />
        </div>

        {status.lastError && (
          <p className="mt-4 rounded-sm bg-red-50 px-3 py-2 text-sm text-danger dark:bg-red-900/20">
            {status.lastError}
          </p>
        )}

        <p className="mt-4 text-sm text-muted-foreground">
          {status.mode === 'agent' ? t('status.agentSummary') : t('status.localSummary')}
        </p>
      </Section>

      <Section icon={<Unlink2Icon size={20} />} title={t('unpair.title')} description={t('unpair.description')}>
        <Button variant="destructive" onClick={onUnpair} icon={<Unlink2Icon size={14} />}>
          {t('unpair.action')}
        </Button>
      </Section>
    </>
  )
}

interface PairFormProps {
  defaultName: string
  onPaired: (status: main.OdooStatus) => void
  onError: (message: string) => void
}

function PairForm({ defaultName, onPaired, onError }: PairFormProps) {
  const { t } = useTranslation('odoo')
  const [pairing, setPairing] = useState('')
  const [name, setName] = useState(defaultName)
  const [mode, setMode] = useState('agent')
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    setName((current) => current || defaultName)
  }, [defaultName])

  async function onSubmit() {
    setIsSaving(true)
    try {
      onPaired(await PairWithOdoo(pairing.trim(), name.trim(), mode))
      setPairing('')
    } catch (err) {
      onError(errorText(err, t('connect.failed')))
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <Section icon={<LinkIcon size={20} />} title={t('connect.title')} description={t('connect.description')}>
      <div className="grid gap-4">
        <Field label={t('connect.token')} hint={t('connect.tokenHint')}>
          <Input
            value={pairing}
            onChange={(e) => setPairing(e.target.value)}
            placeholder="https://odoo.empresa.com?token=…"
            className="font-mono"
          />
        </Field>

        <Field label={t('connect.name')} hint={t('connect.nameHint')}>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>

        <fieldset className="grid gap-2">
          <legend className="mb-1 text-sm font-medium">{t('connect.modeQuestion')}</legend>
          <RadioCard
            checked={mode === 'agent'}
            onChange={() => setMode('agent')}
            title={t('mode.agent')}
            description={t('mode.agentHint')}
          />
          <RadioCard
            checked={mode === 'local'}
            onChange={() => setMode('local')}
            title={t('mode.local')}
            description={t('mode.localHint')}
          />
        </fieldset>

        <div>
          <Button onClick={onSubmit} loading={isSaving} disabled={!pairing.trim()}>
            {t('connect.submit')}
          </Button>
        </div>
      </div>
    </Section>
  )
}
