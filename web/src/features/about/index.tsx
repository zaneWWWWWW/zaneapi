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
import { useTranslation } from 'react-i18next'

import { Footer } from '@/components/layout/components/footer'
import { PublicLayout } from '@/components/layout'
import { SiteDocument } from '@/components/site-document'
import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'

import { getAboutContent } from './api'

function readAboutDocument(raw: unknown): string {
  if (typeof raw === 'string') return raw.trim()
  if (!raw || typeof raw !== 'object') return ''
  const record = raw as Record<string, unknown>
  return ['intro', 'commitments', 'contact']
    .map((key) => {
      const value = record[key]
      return typeof value === 'string' ? value.trim() : ''
    })
    .filter((value) => value.length > 0)
    .join('\n\n')
}

export function About() {
  const { t } = useTranslation()
  const { systemName } = useSystemConfig()
  const { data, isLoading } = useQuery({
    queryKey: ['about-content'],
    queryFn: getAboutContent,
  })

  const content = readAboutDocument(data?.data)
  const hasContent = content.length > 0

  return (
    <PublicLayout showMainContainer={false}>
      <div className='pt-14'>
        <div className='mx-auto flex max-w-3xl flex-col gap-8 px-4 py-10 sm:px-6'>
          <header className='space-y-2'>
            <h1 className='text-2xl font-semibold tracking-tight'>
              {t('About')}
              {systemName ? ` · ${systemName}` : ''}
            </h1>
            <p className='text-muted-foreground text-sm leading-relaxed'>
              {t('Who we are and how to reach us.')}
            </p>
          </header>

          {isLoading ? (
            <div className='flex flex-col gap-3'>
              <Skeleton className='h-8 w-[45%]' />
              <Skeleton className='h-4 w-full' />
              <Skeleton className='h-4 w-[90%]' />
            </div>
          ) : null}

          {!isLoading && !hasContent ? (
            <p className='text-muted-foreground text-sm'>
              {t('The administrator has not added about content yet.')}
            </p>
          ) : null}

          {!isLoading && hasContent ? (
            <SiteDocument content={content} variant='main' />
          ) : null}
        </div>
        <Footer />
      </div>
    </PublicLayout>
  )
}
