import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PlugZapIcon, PrinterIcon, ServerIcon } from 'lucide-react'
import { AppContextWrapper } from '@/contexts/AppContext'
import { PrinterContextWrapper } from '@/contexts/PrinterContext'
import { ToastContextWrapper } from '@/contexts/ToastContext'
import OdooPanel from '@/features/odoo/OdooPanel'
import PrintersPanel from '@/features/printers/PrintersPanel'
import SystemPanel from '@/features/system/SystemPanel'
import AppShell, { type Tab } from '@/layout/AppShell'

function App() {
  const { t } = useTranslation()
  const [tab, setTab] = useState(() => window.location.hash.slice(1) || 'printers')

  // The tab lives in the URL hash, so reloading keeps it and each tab has its own link.
  const changeTab = (id: string) => {
    window.location.hash = id
    setTab(id)
  }

  const tabs: Tab[] = [
    { id: 'printers', label: t('tabs.printers'), icon: PrinterIcon },
    { id: 'odoo', label: t('tabs.odoo'), icon: PlugZapIcon },
    { id: 'system', label: t('tabs.system'), icon: ServerIcon },
  ]

  return (
    <ToastContextWrapper>
      <AppContextWrapper>
        <PrinterContextWrapper>
          <AppShell tabs={tabs} current={tab} onChange={changeTab}>
            {tab === 'printers' && <PrintersPanel />}
            {tab === 'odoo' && <OdooPanel />}
            {tab === 'system' && <SystemPanel />}
          </AppShell>
        </PrinterContextWrapper>
      </AppContextWrapper>
    </ToastContextWrapper>
  )
}

export default App
