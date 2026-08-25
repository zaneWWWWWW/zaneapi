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
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { useApiInfo } from '@/features/dashboard/hooks/use-status-data'
import type { ApiInfoItem } from '@/features/dashboard/types'
import { useStatus } from '@/hooks/use-status'
import { getBgColorClass } from '@/lib/colors'

function trimTrailingSlash(url: string): string {
  return url.replace(/\/+$/, '')
}

function readStatusServerAddress(status: unknown): string {
  if (!status || typeof status !== 'object') return ''
  const record = status as Record<string, unknown>
  const nested =
    record.data && typeof record.data === 'object'
      ? (record.data as Record<string, unknown>)
      : undefined
  const candidate = record.server_address ?? nested?.server_address
  return typeof candidate === 'string' ? trimTrailingSlash(candidate) : ''
}

export function ApiAccessUrls() {
  const { t } = useTranslation()
  const { status } = useStatus()
  const { items, loading } = useApiInfo()

  const endpoints = useMemo(() => {
    const configured = items.filter((item) => item.url?.trim())
    if (configured.length > 0) return configured

    const fallback =
      readStatusServerAddress(status) ||
      (typeof window === 'undefined' ? '' : window.location.origin)
    if (!fallback) return []

    return [
      {
        url: fallback,
        route: t('Direct'),
        description: t('Direct connection'),
        color: 'blue',
      } satisfies ApiInfoItem,
    ]
  }, [items, status, t])

  if (loading || endpoints.length === 0) return null

  return (
    <section className='shrink-0 rounded-lg border px-3 py-2.5 sm:px-4'>
      <div className='mb-2 flex flex-col gap-0.5 sm:flex-row sm:items-baseline sm:justify-between sm:gap-3'>
        <h3 className='text-sm font-medium'>{t('API endpoints')}</h3>
        <p className='text-muted-foreground text-xs'>
          {t('Use the Cloudflare URL if you are in mainland China.')}
        </p>
      </div>
      <ul
        className={
          endpoints.length > 1
            ? 'grid gap-2 sm:grid-cols-2'
            : 'grid gap-2'
        }
      >
        {endpoints.map((item) => (
          <li
            key={`${item.route}-${item.url}`}
            className='bg-muted/30 flex min-w-0 items-center gap-2.5 rounded-md px-2.5 py-2'
          >
            <span
              className={`size-2 shrink-0 rounded-full ${getBgColorClass(item.color)}`}
            />
            <div className='min-w-0 flex-1'>
              <div className='flex items-baseline gap-2'>
                <span className='text-sm font-medium'>{t(item.route)}</span>
                {item.description ? (
                  <span className='text-muted-foreground hidden truncate text-xs sm:inline'>
                    {t(item.description)}
                  </span>
                ) : null}
              </div>
              <p className='text-muted-foreground truncate font-mono text-xs'>
                {item.url}
              </p>
            </div>
            <CopyButton
              value={item.url}
              variant='ghost'
              size='sm'
              className='size-7 p-0'
              iconClassName='size-3.5'
              tooltip={t('Copy URL')}
              aria-label={t('Copy URL')}
            />
          </li>
        ))}
      </ul>
    </section>
  )
}
