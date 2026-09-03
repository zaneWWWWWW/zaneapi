import { useQuery } from '@tanstack/react-query'
import { CalendarDays, CircleDollarSign, RefreshCw } from 'lucide-react'
import { type ReactNode, useMemo, useState } from 'react'
import type { DateRange } from 'react-day-picker'
import { enUS, fr, ja, ru, vi, zhCN, zhTW } from 'react-day-picker/locale'
import { useTranslation } from 'react-i18next'

import {
  sideDrawerContentClassName,
  sideDrawerHeaderClassName,
} from '@/components/drawer-layout'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import dayjs from '@/lib/dayjs'
import { formatPercent, formatQuota } from '@/lib/format'

import { getChannelProfit } from '../api'
import type { ChannelProfitSummary } from '../types'

const calendarLocales = {
  en: enUS,
  zhCN,
  zhTW,
  fr,
  ru,
  ja,
  vi,
} as const

const PROFIT_PRESETS = ['today', 'month', '30d'] as const

type ProfitPreset = (typeof PROFIT_PRESETS)[number] | 'custom'

type ProfitDateRange = {
  start: Date
  end: Date
}

function ProfitMetric(props: { label: string; value: string }) {
  return (
    <div className='border-b p-3 sm:border-r sm:last:border-r-0'>
      <div className='text-muted-foreground text-xs'>{props.label}</div>
      <div className='mt-1 text-base font-medium tabular-nums'>
        {props.value}
      </div>
    </div>
  )
}

function getMargin(summary: ChannelProfitSummary | undefined): string {
  if (!summary || summary.revenue_quota === 0) return '-'
  return formatPercent((summary.profit_quota / summary.revenue_quota) * 100)
}

function getProfitTimeZone(language: string): string {
  switch (language) {
    case 'zhCN':
      return 'Asia/Shanghai'
    case 'zhTW':
      return 'Asia/Taipei'
    case 'ja':
      return 'Asia/Tokyo'
    case 'vi':
      return 'Asia/Ho_Chi_Minh'
    case 'fr':
      return 'Europe/Paris'
    case 'ru':
      return 'Europe/Moscow'
    default:
      return 'UTC'
  }
}

function calendarDateInTimeZone(instant: Date, timeZone: string): Date {
  const zoned = dayjs(instant).tz(timeZone)
  return new Date(zoned.year(), zoned.month(), zoned.date())
}

function instantFromCalendarDate(
  date: Date,
  timeZone: string,
  bound: 'start' | 'end'
): Date {
  const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
  const zoned = dayjs.tz(key, timeZone)
  if (bound === 'start') return zoned.startOf('day').toDate()
  return zoned.endOf('day').toDate()
}

function getProfitPresetRange(
  preset: Exclude<ProfitPreset, 'custom'>,
  nowMs: number,
  timeZone: string
): ProfitDateRange {
  const now = dayjs(nowMs).tz(timeZone)
  if (preset === 'today') {
    return {
      start: now.startOf('day').toDate(),
      end: now.endOf('day').toDate(),
    }
  }
  if (preset === 'month') {
    return {
      start: now.startOf('month').toDate(),
      end: now.endOf('day').toDate(),
    }
  }
  return {
    start: now.subtract(29, 'day').startOf('day').toDate(),
    end: now.endOf('day').toDate(),
  }
}

export function ChannelProfitSheet() {
  const { t, i18n } = useTranslation()
  const timeZone = getProfitTimeZone(i18n.language)
  const [open, setOpen] = useState(false)
  const [openedAt, setOpenedAt] = useState(() => Date.now())
  const [preset, setPreset] = useState<ProfitPreset>('30d')
  const [customRange, setCustomRange] = useState<ProfitDateRange | null>(null)
  const [calendarOpen, setCalendarOpen] = useState(false)
  const range = useMemo(() => {
    if (preset === 'custom' && customRange) return customRange
    if (preset === 'custom') return getProfitPresetRange('30d', openedAt, timeZone)
    return getProfitPresetRange(preset, openedAt, timeZone)
  }, [customRange, openedAt, preset, timeZone])
  const timeRange = useMemo(
    () => ({
      start_timestamp: Math.floor(range.start.getTime() / 1000),
      end_timestamp: Math.floor(range.end.getTime() / 1000),
    }),
    [range]
  )
  const calendarLocale =
    calendarLocales[i18n.language as keyof typeof calendarLocales] ?? enUS
  const currentYear = new Date().getFullYear()
  const [pendingRange, setPendingRange] = useState<DateRange>({
    from: range.start,
    to: range.end,
  })
  const applyCalendarRange = (
    nextRange: DateRange | undefined,
    triggerDate?: Date
  ) => {
    const previousComplete = Boolean(pendingRange.from && pendingRange.to)
    if (previousComplete) {
      setPendingRange({ from: triggerDate ?? nextRange?.from, to: undefined })
      return
    }
    setPendingRange(nextRange ?? { from: undefined })
    if (!nextRange?.from || !nextRange.to) return
    const startDate =
      nextRange.from.getTime() > nextRange.to.getTime()
        ? nextRange.to
        : nextRange.from
    const endDate =
      nextRange.from.getTime() > nextRange.to.getTime()
        ? nextRange.from
        : nextRange.to
    setCustomRange({
      start: instantFromCalendarDate(startDate, timeZone, 'start'),
      end: instantFromCalendarDate(endDate, timeZone, 'end'),
    })
    setPreset('custom')
    setCalendarOpen(false)
  }
  const profitQuery = useQuery({
    queryKey: ['channel-profit', timeRange],
    queryFn: () => getChannelProfit(timeRange),
    enabled: open,
    staleTime: 30 * 1000,
    retry: false,
  })
  const report = profitQuery.data?.data
  let content: ReactNode
  if (profitQuery.isLoading) {
    content = <Skeleton className='h-72 w-full' />
  } else if (profitQuery.isError) {
    content = (
      <p className='text-muted-foreground py-12 text-center text-sm'>
        {t('Unable to load profit data.')}
      </p>
    )
  } else {
    content = (
      <>
        <div className='grid border sm:grid-cols-4'>
          <ProfitMetric
            label={t('Revenue')}
            value={formatQuota(report?.revenue_quota ?? 0)}
          />
          <ProfitMetric
            label={t('Cost')}
            value={formatQuota(report?.cost_quota ?? 0)}
          />
          <ProfitMetric
            label={t('Profit')}
            value={formatQuota(report?.profit_quota ?? 0)}
          />
          <ProfitMetric label={t('Profit Margin')} value={getMargin(report)} />
        </div>

        {report && report.unconfigured_channel_count > 0 ? (
          <p className='text-muted-foreground mt-3 text-sm'>
            {t(
              '{{count}} enabled channels have no cost ratio and are omitted from profit accounting.',
              {
                count: report.unconfigured_channel_count,
              }
            )}
          </p>
        ) : null}

        <div className='mt-4 border'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Channel')}</TableHead>
                <TableHead className='text-right'>{t('Requests')}</TableHead>
                <TableHead className='text-right'>{t('Revenue')}</TableHead>
                <TableHead className='text-right'>{t('Cost')}</TableHead>
                <TableHead className='text-right'>{t('Profit')}</TableHead>
                <TableHead className='text-right'>
                  {t('Profit Margin')}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {report?.channels.map((channel) => (
                <TableRow key={channel.channel_id}>
                  <TableCell className='max-w-48 truncate font-medium'>
                    {channel.channel_name ||
                      t('Deleted channel {{id}}', { id: channel.channel_id })}
                  </TableCell>
                  <TableCell className='text-right'>
                    {channel.request_count}
                  </TableCell>
                  <TableCell className='text-right'>
                    {formatQuota(channel.revenue_quota)}
                  </TableCell>
                  <TableCell className='text-right'>
                    {formatQuota(channel.cost_quota)}
                  </TableCell>
                  <TableCell className='text-right font-medium'>
                    {formatQuota(channel.profit_quota)}
                  </TableCell>
                  <TableCell className='text-right'>
                    {getMargin(channel)}
                  </TableCell>
                </TableRow>
              ))}
              {!report?.channels.length ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className='text-muted-foreground h-32 text-center'
                  >
                    {t('No data')}
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </div>
      </>
    )
  }

  return (
    <>
      <Button variant='outline' onClick={() => setOpen(true)}>
        <CircleDollarSign />
        {t('Profit')}
      </Button>
      <Sheet
        open={open}
        onOpenChange={(nextOpen) => {
          setOpen(nextOpen)
          if (nextOpen) setOpenedAt(Date.now())
        }}
      >
        <SheetContent className={sideDrawerContentClassName('sm:max-w-4xl')}>
          <SheetHeader className={sideDrawerHeaderClassName()}>
            <div className='flex items-center justify-between gap-3 pr-9'>
              <SheetTitle>{t('Channel Profit')}</SheetTitle>
              <Button
                variant='ghost'
                size='icon'
                onClick={() => profitQuery.refetch()}
                aria-label={t('Refresh')}
                disabled={profitQuery.isFetching}
              >
                <RefreshCw
                  className={profitQuery.isFetching ? 'animate-spin' : ''}
                />
              </Button>
            </div>
            <div className='flex flex-wrap items-center gap-1 pt-2'>
              {PROFIT_PRESETS.map((item) => {
                let label = t('Last {{count}} days', { count: 30 })
                if (item === 'today') label = t('Today')
                if (item === 'month') label = t('This month')
                return (
                  <Button
                    key={item}
                    variant={preset === item ? 'secondary' : 'ghost'}
                    size='sm'
                    onClick={() => setPreset(item)}
                  >
                    {label}
                  </Button>
                )
              })}
              <Popover
                open={calendarOpen}
                onOpenChange={(nextOpen) => {
                  setCalendarOpen(nextOpen)
                  if (nextOpen) {
                    setPendingRange({
                      from: calendarDateInTimeZone(range.start, timeZone),
                      to: calendarDateInTimeZone(range.end, timeZone),
                    })
                  }
                }}
              >
                <PopoverTrigger
                  render={
                    <Button
                      variant={preset === 'custom' ? 'secondary' : 'outline'}
                      size='sm'
                      className='font-normal tabular-nums'
                      aria-label={t('Date Range')}
                    />
                  }
                >
                  <CalendarDays />
                  {dayjs(range.start).tz(timeZone).format('YYYY-MM-DD')}
                  {' ~ '}
                  {dayjs(range.end).tz(timeZone).format('YYYY-MM-DD')}
                </PopoverTrigger>
                <PopoverContent align='start' className='w-auto p-0'>
                  <Calendar
                    mode='range'
                    captionLayout='dropdown'
                    locale={calendarLocale}
                    today={calendarDateInTimeZone(new Date(openedAt), timeZone)}
                    selected={pendingRange}
                    startMonth={new Date(currentYear - 10, 0)}
                    endMonth={new Date(currentYear, 11)}
                    disabled={(date) =>
                      date > calendarDateInTimeZone(new Date(openedAt), timeZone)
                    }
                    onSelect={applyCalendarRange}
                  />
                </PopoverContent>
              </Popover>
              <span className='text-muted-foreground px-1 text-xs tabular-nums'>
                UTC{dayjs(openedAt).tz(timeZone).format('Z')}
              </span>
            </div>
          </SheetHeader>

          <div className='min-h-0 flex-1 overflow-y-auto px-4 pb-4'>
            {content}
          </div>
        </SheetContent>
      </Sheet>
    </>
  )
}
