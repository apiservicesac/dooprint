import { useContext, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { CheckIcon, CopyIcon, NetworkIcon, PrinterIcon, Trash2Icon, UsbIcon } from 'lucide-react'
import { main } from '@/../wailsjs/go/models'
import { PrinterContext } from '@/contexts/PrinterContext'
import { ToastContext } from '@/contexts/ToastContext'
import { errorText } from '@/error'
import { executePrint } from '@/functions/executePrint'
import { Badge } from '@/components/ui/badge'
import Button from '@/components/ui/button'
import StatusDot from '@/components/ui/status_dot'

interface PrinterRowProps {
  printer: main.Printer
  status?: 'loading' | 'online' | 'offline'
}

export default function PrinterRow({ printer, status }: PrinterRowProps) {
  const { t } = useTranslation(['printers', 'common'])
  const printerContext = useContext(PrinterContext)
  const toastContext = useContext(ToastContext)
  const [copied, setCopied] = useState(false)
  const [isPrinting, setIsPrinting] = useState(false)

  const tone = status === 'loading' ? 'warning' : status === 'offline' ? 'danger' : 'success'
  const label =
    status === 'loading' ? t('status.checking') : status === 'offline' ? t('status.offline') : t('status.ready')

  async function onCopy() {
    try {
      await navigator.clipboard.writeText(printer.ip)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      toastContext.actions.showToast(errorText(err, t('copyFailed')), 'danger')
    }
  }

  async function onTest() {
    setIsPrinting(true)
    try {
      await executePrint(printer)
      toastContext.actions.showToast(t('testSent', { name: printer.name }), 'success')
    } catch (err) {
      toastContext.actions.showToast(errorText(err, t('testFailed')), 'danger')
    } finally {
      setIsPrinting(false)
    }
  }

  async function onRemove() {
    const result = await printerContext.actions.removeLanPrinter(printer)
    if (result.status) {
      toastContext.actions.showToast(t('removed', { ip: printer.lanIp }), 'success')
    }
  }

  return (
    <li className="py-6 first:pt-1 last:pb-1">
      <div className="flex items-start justify-between gap-6">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2.5">
            <span className="truncate text-base font-medium text-foreground">{printer.name}</span>
            <Badge variant={printer.isLAN ? 'default' : 'muted'}>
              {printer.isLAN ? <NetworkIcon size={12} /> : <UsbIcon size={12} />}
              {printer.isLAN ? t('type.lan') : t('type.usb')}
            </Badge>
            {printer.type === 'label' && <Badge variant="muted">{t('type.label')}</Badge>}
          </div>
          <p className="mt-2 break-all text-sm tabular-nums text-muted-foreground">{printer.ip}</p>
        </div>
        <StatusDot tone={tone} label={label} />
      </div>

      <div className="mt-5 flex flex-wrap gap-3">
        <Button
          variant="outline"
          onClick={onCopy}
          icon={copied ? <CheckIcon size={14} /> : <CopyIcon size={14} />}
        >
          {copied ? t('common:copied') : t('common:copy')}
        </Button>
        <Button variant="outline" onClick={onTest} loading={isPrinting} icon={<PrinterIcon size={14} />}>
          {t('test')}
        </Button>
        {printer.isLAN && (
          <Button variant="outline" onClick={onRemove} icon={<Trash2Icon size={14} />} className="text-red-600! sm:ms-auto">
            {t('common:remove')}
          </Button>
        )}
      </div>
    </li>
  )
}
