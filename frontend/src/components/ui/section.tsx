import type { ReactNode } from 'react'

interface SectionProps {
  title: string
  description?: string
  actions?: ReactNode
  children: ReactNode
}

/** A titled block: the building block of every screen. */
export default function Section({ title, description, actions, children }: SectionProps) {
  return (
    <section className="rounded-lg border border-border bg-card">
      <header className="flex flex-wrap items-start justify-between gap-4 px-6 pt-6 sm:px-8 sm:pt-7">
        <div className="flex min-w-0 items-start">
          <div className="min-w-0">
            <h2 className="text-lg font-semibold leading-7 text-foreground">{title}</h2>
            {description && <p className="mt-1 text-[15px] leading-6 text-muted-foreground">{description}</p>}
          </div>
        </div>
        {actions && <div className="flex shrink-0 items-center gap-3">{actions}</div>}
      </header>
      <div className="px-6 pb-6 pt-5 sm:px-8 sm:pb-8">{children}</div>
    </section>
  )
}
