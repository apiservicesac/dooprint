import * as React from 'react'
import { useTranslation } from 'react-i18next'
import { CheckIcon, CopyIcon, XIcon } from 'lucide-react'
import legacyCopy from 'copy-text-to-clipboard'
import { cn } from '@/lib/utils'

interface Props {
  text: string
}

type State = 'idle' | 'copied' | 'failed'

/**
 * The async clipboard API first, the textarea trick only as a fallback.
 *
 * The old path works by selecting a hidden textarea and asking the document to copy
 * the selection, which needs focus it does not always get: inside a dialog the focus
 * trap takes it straight back, the selection is gone by the time the copy runs, and
 * nothing lands on the clipboard. `navigator.clipboard` needs no selection at all.
 *
 * It is still needed, though — the API is only available over HTTPS and localhost.
 */
export default function CopyButton({ text }: Props) {
  const { t } = useTranslation()
  const [state, setState] = React.useState<State>('idle')

  const copy = async () => {
    // Saying "copied" when nothing was copied is the worst of the three answers:
    // this one shows a token that cannot be read again.
    setState(await write(text) ? 'copied' : 'failed')
    setTimeout(() => setState('idle'), 2000)
  }

  return (
    <button
      type="button"
      onClick={copy}
      className={cn(
        'inline-flex shrink-0 items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-semibold transition-colors',
        state === 'failed'
          ? 'border-red-300 bg-card text-red-600 dark:border-red-900 dark:bg-zinc-800 dark:text-red-400'
          : 'border-border bg-card text-foreground hover:border-zinc-400 hover:bg-muted dark:border-zinc-600 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:border-zinc-500 dark:hover:bg-zinc-700',
      )}
    >
      {state === 'copied' && <CheckIcon className="h-3.5 w-3.5 text-brand-teal" />}
      {state === 'failed' && <XIcon className="h-3.5 w-3.5" />}
      {state === 'idle'   && <CopyIcon className="h-3.5 w-3.5" />}
      {t(state === 'copied' ? 'copied' : state === 'failed' ? 'copyFailed' : 'copy')}
    </button>
  )
}

async function write(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return legacyCopy(text)
  }
}
