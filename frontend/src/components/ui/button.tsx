import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import Spinner from '@/components/ui/spinner'
import { cn } from '@/lib/utils'

const buttonVariants = cva('', {
  variants: {
    variant: {
      default:     'btn-primary',
      outline:     'btn-secondary',
      destructive: 'btn-red',
      ghost:       'hover:bg-accent hover:text-accent-foreground',
    },
  },
  defaultVariants: { variant: 'default' },
})

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> {
  loading?: boolean
  icon?: React.ReactNode
}

export default function Button({ children, disabled, loading, variant, className, icon, ...props }: ButtonProps) {
  return (
    <button disabled={loading || disabled} {...props} className={cn(buttonVariants({ variant }), className)}>
      {loading ? <Spinner className="h-4 w-4 shrink-0 animate-spin" /> : icon}
      {/* An icon-only button gets no empty label: the flex gap would pad it off-centre. */}
      {children != null && children !== false && <span className="whitespace-nowrap">{children}</span>}
    </button>
  )
}
