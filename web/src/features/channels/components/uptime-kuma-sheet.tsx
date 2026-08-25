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
import { RotateCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import {
  sideDrawerContentClassName,
  sideDrawerFormClassName,
  sideDrawerHeaderClassName,
} from '@/components/drawer-layout'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  getOptionValue,
  useSystemOptions,
} from '@/features/system-settings/hooks/use-system-options'
import { cn } from '@/lib/utils'

import { getUptimeStatus } from '../api'
import {
  flattenUptimeMonitors,
  listUnmatchedMonitors,
  normalizeProbeName,
} from '../lib/uptime-match'
import type { UptimeMonitor } from '../types'
import { ChannelProbeCell } from './channel-probe-cell'
import { UptimeKumaSettings } from './uptime-kuma-settings'

const KUMA_OPTION_DEFAULTS = {
  'console_setting.uptime_kuma_enabled': false,
  'console_setting.uptime_kuma_groups': '[]',
}

type UptimeKumaSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  channelNames: string[]
}

function MonitorRow(props: { monitor: UptimeMonitor }) {
  return (
    <div className='flex items-center justify-between gap-3 border-b px-1 py-2 last:border-b-0'>
      <div className='min-w-0'>
        <div className='truncate text-sm'>{props.monitor.name}</div>
        {props.monitor.group ? (
          <div className='text-muted-foreground truncate text-xs'>
            {props.monitor.group}
          </div>
        ) : null}
      </div>
      <ChannelProbeCell monitor={props.monitor} />
    </div>
  )
}

export function UptimeKumaSheet(props: UptimeKumaSheetProps) {
  const { t } = useTranslation()
  const optionsQuery = useSystemOptions()
  const settings = getOptionValue(
    optionsQuery.data?.data,
    KUMA_OPTION_DEFAULTS
  )
  const statusQuery = useQuery({
    queryKey: ['uptime-kuma-status'],
    queryFn: getUptimeStatus,
    enabled: props.open && settings['console_setting.uptime_kuma_enabled'],
    staleTime: 60 * 1000,
    retry: false,
  })

  const groups = statusQuery.data?.data ?? []
  const matchedNames = new Set(props.channelNames.map(normalizeProbeName))
  const allMonitors = flattenUptimeMonitors(groups)
  const matchedMonitors = allMonitors.filter((monitor) =>
    matchedNames.has(normalizeProbeName(monitor.name))
  )
  const unmatchedMonitors = listUnmatchedMonitors(groups, props.channelNames)

  let statusContent
  if (!settings['console_setting.uptime_kuma_enabled']) {
    statusContent = (
      <p className='text-muted-foreground text-sm'>
        {t('Enable Uptime Kuma in Settings to load probes.')}
      </p>
    )
  } else if (statusQuery.isLoading) {
    statusContent = (
      <p className='text-muted-foreground text-sm'>{t('Loading...')}</p>
    )
  } else if (allMonitors.length === 0) {
    statusContent = (
      <p className='text-muted-foreground text-sm'>
        {t('No uptime monitoring configured')}
      </p>
    )
  } else {
    statusContent = (
      <ScrollArea className='h-[min(28rem,50vh)]'>
        {matchedMonitors.length > 0 ? (
          <div className='mb-4'>
            <h4 className='text-muted-foreground mb-1 text-xs font-semibold tracking-wider uppercase'>
              {t('Matched channels')}
            </h4>
            {matchedMonitors.map((monitor) => (
              <MonitorRow key={monitor.name} monitor={monitor} />
            ))}
          </div>
        ) : null}
        {unmatchedMonitors.length > 0 ? (
          <div>
            <h4 className='text-muted-foreground mb-1 text-xs font-semibold tracking-wider uppercase'>
              {t('External probes')}
            </h4>
            {unmatchedMonitors.map((monitor) => (
              <MonitorRow key={monitor.name} monitor={monitor} />
            ))}
          </div>
        ) : null}
      </ScrollArea>
    )
  }

  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent className={sideDrawerContentClassName('sm:max-w-2xl')}>
        <SheetHeader className={sideDrawerHeaderClassName()}>
          <SheetTitle>{t('Uptime Kuma')}</SheetTitle>
          <SheetDescription>
            {t(
              'Name Kuma monitors after channel names to show 24h uptime on each row.'
            )}
          </SheetDescription>
        </SheetHeader>
        <div className={cn(sideDrawerFormClassName(), 'pt-3')}>
          <Tabs defaultValue='status'>
            <TabsList>
              <TabsTrigger value='status'>{t('Status')}</TabsTrigger>
              <TabsTrigger value='settings'>{t('Settings')}</TabsTrigger>
            </TabsList>
            <TabsContent value='status' className='mt-4'>
              <div className='mb-3 flex justify-end'>
                <Button
                  type='button'
                  variant='ghost'
                  size='sm'
                  onClick={() => {
                    void statusQuery.refetch()
                  }}
                  disabled={statusQuery.isFetching}
                >
                  <RotateCw
                    className={cn(
                      'size-3.5',
                      statusQuery.isFetching && 'animate-spin'
                    )}
                  />
                  {t('Refresh')}
                </Button>
              </div>
              {statusContent}
            </TabsContent>
            <TabsContent value='settings' className='mt-4'>
              <UptimeKumaSettings
                enabled={settings['console_setting.uptime_kuma_enabled']}
                data={settings['console_setting.uptime_kuma_groups']}
              />
            </TabsContent>
          </Tabs>
        </div>
      </SheetContent>
    </Sheet>
  )
}
