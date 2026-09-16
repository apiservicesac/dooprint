import type { ReactNode } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface SectionProps {
  icon?: ReactNode
  title: string
  description?: string
  actions?: ReactNode
  children: ReactNode
}

/** Card with a header: the building block of every screen. */
export default function Section({ icon, title, description, actions, children }: SectionProps) {
  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between gap-4">
        <div className="flex min-w-0 items-start gap-3">
          {icon && <span className="mt-1 shrink-0 text-brand-orange">{icon}</span>}
          <div className="min-w-0">
            <CardTitle className="text-xl">{title}</CardTitle>
            {description && <CardDescription className="mt-1">{description}</CardDescription>}
          </div>
        </div>
        {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  )
}
