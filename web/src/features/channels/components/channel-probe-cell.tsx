/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useTranslation } from 'react-i18next'

import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

import type { UptimeMonitor } from '../types'

const STATUS_COLOR_MAP: Record<number, string> = {
  1: 'bg-emerald-500',
  0: 'bg-red-500',
  2: 'bg-amber-500',
  3: 'bg-blue-500',
}

export function ChannelProbeCell(props: { monitor: UptimeMonitor | null }) {
  const { t } = useTranslation()
  if (!props.monitor) {
    return <span className='text-muted-foreground text-xs'>—</span>
  }

  const percent = ((props.monitor.uptime ?? 0) * 100).toFixed(2)
  const color =
    STATUS_COLOR_MAP[props.monitor.status] ?? 'bg-muted-foreground/40'

  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <span className='inline-flex items-center gap-1.5' />
        }
      >
        <span className={cn('size-2 rounded-full', color)} aria-hidden='true' />
        <span className='font-mono text-xs font-medium tabular-nums'>
          {percent}%
        </span>
      </TooltipTrigger>
      <TooltipContent>
        <p>
          {props.monitor.name}
          {props.monitor.group ? ` (${props.monitor.group})` : ''}
        </p>
        <p className='text-muted-foreground'>
          {t('24h')} {percent}%
        </p>
      </TooltipContent>
    </Tooltip>
  )
}
