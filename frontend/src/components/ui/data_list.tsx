import type { ReactNode } from 'react'

interface DataListProps {
  rows: [string, ReactNode][]
}

/** Label/value list for technical details. A value can be text or a badge. */
export default function DataList({ rows }: DataListProps) {
  return (
    <dl className="divide-y divide-border">
      {rows.map(([label, value]) => (
        <div key={label} className="flex items-center justify-between gap-6 py-3.5 first:pt-0 last:pb-0">
          <dt className="text-[15px] text-muted-foreground">{label}</dt>
          <dd className="break-all text-right text-[15px] font-medium text-foreground">
            {value === '' || value === null || value === undefined ? '—' : value}
          </dd>
        </div>
      ))}
    </dl>
  )
}
