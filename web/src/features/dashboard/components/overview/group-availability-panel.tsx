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
import { useQuery } from '@tanstack/react-query'
import { Layers } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { IconBadge } from '@/components/ui/icon-badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  bucketSecondsToMinutes,
  isHourAvailabilityWindow,
} from '@/features/dashboard/lib/group-availability'
import { getPerfMetricsGroups } from '@/features/performance-metrics/api'
import {
  formatUptimePct,
  getSuccessRateDotClass,
  getSuccessRateTextClass,
} from '@/features/performance-metrics/lib/format'
import type { GroupAvailability } from '@/features/performance-metrics/types'
import { cn } from '@/lib/utils'

import { PanelWrapper } from '../ui/panel-wrapper'

function formatRate(rate: number | null | undefined): string {
  if (rate == null) return '—'
  return formatUptimePct(rate)
}

function GroupAvailabilityRow(props: { item: GroupAvailability }) {
  const current = props.item.current_success_rate
  const hours = props.item.hours_success_rate

  return (
    <div className='hover:bg-muted/40 flex items-center justify-between gap-3 px-3 py-2.5 sm:px-5'>
      <div className='flex min-w-0 items-start gap-2.5'>
        <span
          className={cn(
            'mt-1.5 inline-block size-2 shrink-0 rounded-full',
            getSuccessRateDotClass(hours ?? Number.NaN)
          )}
          aria-hidden='true'
        />
        <div className='min-w-0'>
          <div className='truncate font-mono text-sm font-semibold'>
            {props.item.group}
          </div>
          {props.item.description ? (
            <div className='text-muted-foreground/60 truncate text-xs'>
              {props.item.description}
            </div>
          ) : null}
        </div>
      </div>
      <div className='flex shrink-0 items-baseline gap-3 font-mono text-sm tabular-nums'>
        <span
          className={cn(
            'w-16 text-right font-semibold',
            current == null
              ? 'text-muted-foreground'
              : getSuccessRateTextClass(current)
          )}
        >
          {formatRate(current)}
        </span>
        <span
          className={cn(
            'w-14 text-right font-semibold',
            hours == null
              ? 'text-muted-foreground'
              : getSuccessRateTextClass(hours)
          )}
        >
          {formatRate(hours)}
        </span>
      </div>
    </div>
  )
}

export function GroupAvailabilityPanel(props: { fill?: boolean }) {
  const { t } = useTranslation()
  const query = useQuery({
    queryKey: ['perf-metrics-groups', 24],
    queryFn: () => getPerfMetricsGroups(24),
    staleTime: 60 * 1000,
    retry: false,
  })

  const groups = query.data?.data?.groups ?? []
  const minutes = bucketSecondsToMinutes(query.data?.data?.bucket_seconds)
  const hourWindow = isHourAvailabilityWindow(minutes)
  const loading = query.isLoading
  const empty = !loading && groups.length === 0
  const description = hourWindow
    ? t('Success rate in the last 1 hour and 24 hours')
    : t('Success rate in the last {{minutes}} minutes and 24 hours', {
        minutes,
      })
  const currentLabel = hourWindow ? t('1h') : t('{{minutes}} min', { minutes })

  return (
    <PanelWrapper
      title={
        <span className='flex items-center gap-2'>
          <IconBadge tone='success' size='sm'>
            <Layers />
          </IconBadge>
          {t('Group availability')}
        </span>
      }
      description={description}
      loading={loading}
      empty={empty}
      emptyMessage={t('No groups available')}
      height='h-80'
      fill={props.fill}
      contentClassName='p-0'
    >
      <ScrollArea className={props.fill ? 'h-full min-h-0 flex-1' : 'h-80'}>
        <div
          className='text-muted-foreground/70 border-border/60 grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 border-b px-3 py-2 text-[11px] font-medium tracking-wide uppercase sm:px-5'
          aria-hidden='true'
        >
          <span>{t('Group')}</span>
          <div className='flex gap-3'>
            <span className='w-16 text-right'>{currentLabel}</span>
            <span className='w-14 text-right'>{t('24h')}</span>
          </div>
        </div>
        <div>
          {groups.map((item) => (
            <GroupAvailabilityRow key={item.group} item={item} />
          ))}
        </div>
      </ScrollArea>
    </PanelWrapper>
  )
}
