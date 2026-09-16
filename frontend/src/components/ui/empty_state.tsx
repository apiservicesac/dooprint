import type { ReactNode } from 'react'
import { PrinterIcon } from 'lucide-react'

interface EmptyStateProps {
  icon?: ReactNode
  title: string
  description?: string
  action?: ReactNode
}

export default function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="px-4 py-12 text-center">
      <div className="mb-4 flex justify-center text-muted-foreground/60">
        {icon ?? <PrinterIcon size={36} strokeWidth={1.5} />}
      </div>
      <p className="text-base font-medium text-foreground">{title}</p>
      {description && <p className="mx-auto mt-2 max-w-sm text-[15px] leading-6 text-muted-foreground">{description}</p>}
      {action && <div className="mt-6 flex justify-center">{action}</div>}
    </div>
  )
}
