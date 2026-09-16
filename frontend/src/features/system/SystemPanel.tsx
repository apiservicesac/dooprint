import { useContext, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  AppleIcon,
  FolderIcon,
  MonitorIcon,
  ServerIcon,
  ShieldCheckIcon,
  ShieldOffIcon,
  TerminalIcon,
  WifiIcon,
} from 'lucide-react'
import { GetTroubleshootInfo } from '@/../wailsjs/go/main/App'
import { main } from '@/../wailsjs/go/models'
import { AppContext } from '@/contexts/AppContext'
import { Badge } from '@/components/ui/badge'
import DataList from '@/components/ui/data_list'
import Section from '@/components/ui/section'

const OS_ICONS = {
  linux: TerminalIcon,
  windows: MonitorIcon,
  darwin: AppleIcon,
} as const

export default function SystemPanel() {
  const { t } = useTranslation(['system', 'common'])
  const { app } = useContext(AppContext).data
  const [info, setInfo] = useState<main.TroubleshootInfo | null>(null)

  useEffect(() => {
    GetTroubleshootInfo()
      .then(setInfo)
      .catch((err) => console.error('Cannot read system information', err))
  }, [])

  const os = (app?.os ?? '') as keyof typeof OS_ICONS
  const OsIcon = OS_ICONS[os] ?? ServerIcon
  const firewall = info?.activeFirewall

  return (
    <>
      <Section icon={<ServerIcon size={20} />} title={t('device.title')}>
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

      <Section icon={<WifiIcon size={20} />} title={t('network.title')} description={t('network.description')}>
        <DataList
          rows={[
            [t('network.address'), <span className="font-mono text-xs">{app?.address}</span>],
            [t('network.port'), app ? String(app.port) : ''],
            [t('network.localIp'), <span className="font-mono text-xs">{info?.localIp}</span>],
            [t('network.subnet'), <span className="font-mono text-xs">{info?.subnet}</span>],
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

      <Section icon={<FolderIcon size={20} />} title={t('install.title')}>
        <DataList
          columns={1}
          rows={[[t('install.path'), <span className="font-mono text-xs">{info?.execPath}</span>]]}
        />
      </Section>
    </>
  )
}
