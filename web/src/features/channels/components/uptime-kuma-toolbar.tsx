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
import { Activity } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { useStatus } from '@/hooks/use-status'

import { getUptimeStatus } from '../api'
import {
  flattenUptimeMonitors,
  summarizeProbes,
} from '../lib/uptime-match'
import { UptimeKumaSheet } from './uptime-kuma-sheet'

export function UptimeKumaToolbar(props: { channelNames: string[] }) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const enabled = Boolean(status) && status?.uptime_kuma_enabled !== false
  const [open, setOpen] = useState(false)
  const statusQuery = useQuery({
    queryKey: ['uptime-kuma-status'],
    queryFn: getUptimeStatus,
    enabled,
    staleTime: 60 * 1000,
    retry: false,
  })
  const summary = summarizeProbes(
    flattenUptimeMonitors(statusQuery.data?.data)
  )

  let label = t('Probes')
  if (enabled && summary.total > 0) {
    label = `${summary.up}/${summary.total}`
  }

  return (
    <>
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              type='button'
              variant='outline'
              size='sm'
              onClick={() => setOpen(true)}
              aria-label={t('Uptime Kuma')}
            />
          }
        >
          <Activity />
          <span className='hidden sm:inline'>{label}</span>
        </TooltipTrigger>
        <TooltipContent>
          {enabled
            ? t('Channel probes from Uptime Kuma')
            : t('Configure Uptime Kuma')}
        </TooltipContent>
      </Tooltip>
      <UptimeKumaSheet
        open={open}
        onOpenChange={setOpen}
        channelNames={props.channelNames}
      />
    </>
  )
}
