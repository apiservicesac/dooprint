import { useContext, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PlusIcon, PrinterIcon } from 'lucide-react'
import NetworkIpDialog from '@/components/NetworkIpDialog'
import { PrinterContext } from '@/contexts/PrinterContext'
import Button from '@/components/ui/button'
import EmptyState from '@/components/ui/empty_state'
import Section from '@/components/ui/section'
import PrinterRow from './PrinterRow'

export default function PrintersPanel() {
  const { t } = useTranslation('printers')
  const { printers, lanStatus, fetchError } = useContext(PrinterContext).data
  const [addOpen, setAddOpen] = useState(0)

  const errorMessage = fetchError ?? printers?.errorMsg
  const available = printers?.printers ?? []
  const unavailable = printers?.unavailablePrinters ?? []
  const addButton = (
    <Button icon={<PlusIcon size={14} />} onClick={() => setAddOpen((count) => count + 1)}>
      {t('add')}
    </Button>
  )

  return (
    <>
      <Section
        icon={<PrinterIcon size={20} />}
        title={t('title')}
        description={t('description')}
        actions={addButton}
      >
        {!printers && !fetchError && <p className="text-sm text-muted-foreground">{t('searching')}</p>}

        {printers && available.length === 0 && unavailable.length === 0 && (
          <EmptyState title={t('empty.title')} description={t('empty.description')} action={addButton} />
        )}

        {(available.length > 0 || unavailable.length > 0) && (
          <ul className="divide-y divide-border">
            {available.map((printer) => (
              <PrinterRow
                key={printer.id}
                printer={printer}
                status={printer.isLAN && printer.lanIp ? lanStatus[printer.lanIp] : 'online'}
              />
            ))}
            {unavailable.map((printer) => (
              <li key={printer.name} className="py-4">
                <p className="font-medium">{printer.name}</p>
                <p className="mt-1 text-sm text-danger">{printer.errorMsg}</p>
              </li>
            ))}
          </ul>
        )}

        {errorMessage && (
          <p className="mt-4 rounded-sm bg-red-50 px-3 py-2 text-sm text-danger dark:bg-red-900/20">{errorMessage}</p>
        )}
      </Section>

      <NetworkIpDialog openSignal={addOpen} />
    </>
  )
}
