import { CircleCheckIcon, TriangleAlertIcon, XIcon } from 'lucide-react'
import type { ToastType } from '@/types'
import { cn } from '@/lib/utils'

export interface ToastItem {
  id: number
  message: string
  type: ToastType
}

const STYLES: Record<ToastType, string> = {
  success: 'border-brand-teal/30 bg-brand-teal-light text-brand-teal-dark dark:text-brand-teal',
  danger: 'border-red-500/30 bg-red-50 text-red-900 dark:bg-red-900/20 dark:text-red-300',
}

const ICONS: Record<ToastType, typeof CircleCheckIcon> = {
  success: CircleCheckIcon,
  danger: TriangleAlertIcon,
}

/** Short notice: shows at the top, goes away on its own and can be closed. */
export default function Toast({ toast, onClose }: { toast: ToastItem; onClose: () => void }) {
  const Icon = ICONS[toast.type]
  return (
    <div
      role="status"
      className={cn(
        'pointer-events-auto flex w-full items-start gap-3 rounded-sm border p-4 pr-10 shadow-lg relative',
        'animate-in slide-in-from-top-4 fade-in duration-200',
        STYLES[toast.type],
      )}
    >
      <Icon size={18} className="mt-0.5 shrink-0" />
      <p className="text-sm leading-snug">{toast.message}</p>
      <button
        type="button"
        onClick={onClose}
        className="absolute right-2 top-2 cursor-pointer rounded-sm p-1 opacity-60 transition-opacity hover:opacity-100"
        aria-label="Cerrar"
      >
        <XIcon size={14} />
      </button>
    </div>
  )
}

export function ToastViewport({ children }: { children: React.ReactNode }) {
  return (
    <div className="pointer-events-none fixed inset-x-0 top-0 z-50 mx-auto flex w-full max-w-md flex-col gap-2 p-4">
      {children}
    </div>
  )
}
