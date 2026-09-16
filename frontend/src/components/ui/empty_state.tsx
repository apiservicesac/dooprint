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
    <div className="py-10 text-center">
      <div className="mb-3 flex justify-center text-muted-foreground/50">
        {icon ?? <PrinterIcon size={40} strokeWidth={1.5} />}
      </div>
      <p className="font-medium text-foreground">{title}</p>
      {description && <p className="mx-auto mt-1 max-w-sm text-sm text-muted-foreground">{description}</p>}
      {action && <div className="mt-4 flex justify-center">{action}</div>}
    </div>
  )
}
