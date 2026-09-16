import { useContext, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { MoonIcon, SunIcon, type LucideIcon } from 'lucide-react'
import { AppContext } from '@/contexts/AppContext'
import { LANGUAGES } from '@/lib/i18n'
import { applyTheme, readTheme, type Theme } from '@/lib/theme'
import { cn } from '@/lib/utils'
import StatusDot from '@/components/ui/status_dot'

export interface Tab {
  id: string
  label: string
  icon: LucideIcon
}

interface AppShellProps {
  tabs: Tab[]
  current: string
  onChange: (id: string) => void
  children: ReactNode
}

/** Application frame: header with device status, language, theme and sections. */
export default function AppShell({ tabs, current, onChange, children }: AppShellProps) {
  const { t, i18n } = useTranslation()
  const { app, serverIsRunning } = useContext(AppContext).data
  const [theme, setTheme] = useState<Theme>(readTheme)

  const toggleTheme = () => {
    const next: Theme = theme === 'dark' ? 'light' : 'dark'
    applyTheme(next)
    setTheme(next)
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border bg-card">
        <div className="mx-auto max-w-4xl px-5 sm:px-8">
          <div className="flex items-center justify-between gap-6 py-6">
            <div className="flex min-w-0 items-center gap-3">
              <img src="/favicon.svg" alt="" className="h-11 w-11 shrink-0" />
              <div className="min-w-0">
                <h1 className="text-lg font-semibold leading-7">{t('appName')}</h1>
                <p className="truncate text-[15px] text-muted-foreground">
                  {app?.name || t('device')}
                  {app?.address ? ` · ${app.address}` : ''}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <StatusDot
                tone={serverIsRunning ? 'success' : 'danger'}
                label={serverIsRunning ? t('running') : t('stopped')}
                className="hidden sm:inline-flex"
              />

              <div className="flex rounded-md border border-border p-1" role="group" aria-label={t('language')}>
                {LANGUAGES.map(({ code, short, label }) => (
                  <button
                    key={code}
                    type="button"
                    title={label}
                    onClick={() => i18n.changeLanguage(code)}
                    className={cn(
                      'cursor-pointer rounded px-2.5 py-1 text-sm font-medium transition-colors',
                      i18n.resolvedLanguage === code
                        ? 'bg-brand-orange text-white'
                        : 'text-muted-foreground hover:text-foreground',
                    )}
                  >
                    {short}
                  </button>
                ))}
              </div>

              <button
                type="button"
                onClick={toggleTheme}
                title={theme === 'dark' ? t('lightMode') : t('darkMode')}
                className="cursor-pointer rounded-md border border-border p-2.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
              >
                {theme === 'dark' ? <SunIcon size={16} /> : <MoonIcon size={16} />}
              </button>
            </div>
          </div>

          <nav className="-mb-px flex gap-8">
            {tabs.map(({ id, label, icon: Icon }) => (
              <button
                key={id}
                onClick={() => onChange(id)}
                className={cn(
                  'flex cursor-pointer items-center gap-2 border-b-2 pb-4 text-[15px] font-medium transition-colors',
                  current === id
                    ? 'border-brand-orange text-foreground'
                    : 'border-transparent text-muted-foreground hover:text-foreground',
                )}
              >
                <Icon size={18} />
                {label}
              </button>
            ))}
          </nav>
        </div>
      </header>

      <main className="mx-auto flex max-w-4xl flex-col gap-8 px-5 py-10 sm:px-8">{children}</main>
    </div>
  )
}
