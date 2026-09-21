import { useContext, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AppleIcon, MonitorIcon, RefreshCwIcon, ServerIcon, ShieldCheckIcon, ShieldOffIcon, TerminalIcon } from 'lucide-react'
import { GetTroubleshootInfo } from '@/../wailsjs/go/main/App'
import { main } from '@/../wailsjs/go/models'
import { AppContext } from '@/contexts/AppContext'
import { ToastContext } from '@/contexts/ToastContext'
import { UpdateContext } from '@/contexts/UpdateContext'
import { Badge } from '@/components/ui/badge'
import Button from '@/components/ui/button'
import DataList from '@/components/ui/data_list'
import Section from '@/components/ui/section'
import { errorText } from '@/error'

const OS_ICONS = {
  linux: TerminalIcon,
  windows: MonitorIcon,
  darwin: AppleIcon,
} as const

export default function SystemPanel() {
  const { t } = useTranslation(['system', 'common'])
  const { app } = useContext(AppContext).data
  const { status, checking } = useContext(UpdateContext).data
  const { check } = useContext(UpdateContext).actions
  const { showToast } = useContext(ToastContext).actions
  const [info, setInfo] = useState<main.TroubleshootInfo | null>(null)

  useEffect(() => {
    GetTroubleshootInfo()
      .then(setInfo)
      .catch((err) => console.error('Cannot read system information', err))
  }, [])

  // The device only looks for a new version when this button is pressed.
  const lookForUpdate = async () => {
    try {
      const result = await check()
      if (result?.error) {
        showToast(result.error, 'danger')
      } else if (!result?.available) {
        showToast(t('update.upToDate'), 'success')
      }
    } catch (error) {
      showToast(errorText(error, t('update.failed')), 'danger')
    }
  }

  const os = (app?.os ?? '') as keyof typeof OS_ICONS
  const OsIcon = OS_ICONS[os] ?? ServerIcon
  const firewall = info?.activeFirewall

  return (
    <>
      <Section title={t('device.title')}>
        <DataList
          rows={[
            [t('device.name'), app?.name ?? ''],
            [t('device.version'), app?.version ?? ''],
            [
              t('device.os'),
              app?.os ? (
                <Badge variant="muted">
                  <OsIcon size={12} />
                  {t(`os.${os}`, { defaultValue: app.os })}
                </Badge>
              ) : (
                ''
              ),
            ],
            [
              t('device.server'),
              <Badge variant={app?.serverRunning ? 'success' : 'danger'}>
                {app?.serverRunning ? t('common:running') : t('common:stopped')}
              </Badge>,
            ],
          ]}
        />
      </Section>

      <Section title={t('network.title')} description={t('network.description')}>
        <DataList
          rows={[
            [t('network.address'), <span className="text-sm tabular-nums">{app?.address}</span>],
            [t('network.port'), app ? String(app.port) : ''],
            [t('network.localIp'), <span className="text-sm tabular-nums">{info?.localIp}</span>],
            [t('network.subnet'), <span className="text-sm tabular-nums">{info?.subnet}</span>],
            [
              t('network.firewall'),
              <Badge variant={firewall ? 'warning' : 'muted'}>
                {firewall ? <ShieldCheckIcon size={12} /> : <ShieldOffIcon size={12} />}
                {firewall || t('network.noFirewall')}
              </Badge>,
            ],
            [t('network.zone'), info?.firewallZone ?? ''],
          ]}
        />
      </Section>

      <Section title={t('update.title')} description={t('update.description')}>
        <DataList
          rows={[
            [t('update.installed'), app?.version ?? ''],
            [
              t('update.latest'),
              status?.latest ? (
                <Badge variant={status.available ? 'warning' : 'success'}>
                  {status.latest}
                  {status.available ? ` · ${t('update.isNewer')}` : ` · ${t('update.upToDate')}`}
                </Badge>
              ) : (
                <span className="text-muted-foreground">{t('update.unknown')}</span>
              ),
            ],
          ]}
        />
        <div className="mt-6">
          <Button variant="outline" onClick={lookForUpdate} loading={checking} icon={<RefreshCwIcon size={16} />}>
            {t('update.check')}
          </Button>
        </div>
      </Section>

      <Section title={t('install.title')}>
        <DataList
          rows={[[t('install.path'), <span className="text-sm tabular-nums">{info?.execPath}</span>]]}
        />
      </Section>
    </>
  )
}
