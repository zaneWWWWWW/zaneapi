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
import { Link } from '@tanstack/react-router'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { Footer } from '@/components/layout/components/footer'
import { PublicLayout } from '@/components/layout'
import { SiteDocument } from '@/components/site-document'
import { Skeleton } from '@/components/ui/skeleton'
import { useApiInfo } from '@/features/dashboard/hooks/use-status-data'
import type { ApiInfoItem } from '@/features/dashboard/types'
import { useStatus } from '@/hooks/use-status'
import { getBgColorClass } from '@/lib/colors'
import { useAuthStore } from '@/stores/auth-store'

import { getDocsContent } from './api'

function trimTrailingSlash(url: string): string {
  return url.replace(/\/+$/, '')
}

function readStatusString(status: unknown, key: 'server_address'): string {
  if (!status || typeof status !== 'object') return ''
  const record = status as Record<string, unknown>
  const nested =
    record.data && typeof record.data === 'object'
      ? (record.data as Record<string, unknown>)
      : undefined
  const candidate = record[key] ?? nested?.[key]
  return typeof candidate === 'string' ? candidate.trim() : ''
}

function curlSample(baseUrl: string, path: string, body: string): string {
  return `curl ${baseUrl}${path} \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '${body}'`
}

type ProtocolGuide = {
  id: string
  titleKey: string
  path: string
  body: string
}

const PROTOCOL_GUIDES: ProtocolGuide[] = [
  {
    id: 'chat',
    titleKey: 'Chat completions',
    path: '/v1/chat/completions',
    body: '{"model":"your-model","messages":[{"role":"user","content":"Hello"}]}',
  },
  {
    id: 'responses',
    titleKey: 'OpenAI Responses',
    path: '/v1/responses',
    body: '{"model":"your-model","input":"Hello"}',
  },
  {
    id: 'anthropic',
    titleKey: 'Anthropic Messages',
    path: '/v1/messages',
    body: '{"model":"your-model","max_tokens":1024,"messages":[{"role":"user","content":"Hello"}]}',
  },
  {
    id: 'gemini',
    titleKey: 'Gemini',
    path: '/v1beta/models/your-model:generateContent',
    body: '{"contents":[{"parts":[{"text":"Hello"}]}]}',
  },
  {
    id: 'image',
    titleKey: 'Image Generation',
    path: '/v1/images/generations',
    body: '{"model":"your-model","prompt":"A red circle"}',
  },
  {
    id: 'video',
    titleKey: 'Video generation',
    path: '/v1/videos',
    body: '{"model":"your-model","prompt":"A cat walking"}',
  },
  {
    id: 'embeddings',
    titleKey: 'Embeddings',
    path: '/v1/embeddings',
    body: '{"model":"your-model","input":"Hello"}',
  },
]

export function Docs() {
  const { t } = useTranslation()
  const { status } = useStatus()
  const { items } = useApiInfo()
  const { data, isLoading } = useQuery({
    queryKey: ['docs-content'],
    queryFn: getDocsContent,
  })
  const isAuthenticated = Boolean(useAuthStore((state) => state.auth.user))
  const customContent =
    typeof data?.data === 'string' ? data.data.trim() : ''
  const hasCustomContent = customContent.length > 0
  const fallbackBase =
    trimTrailingSlash(readStatusString(status, 'server_address')) ||
    (typeof window === 'undefined' ? '' : window.location.origin)

  const accessUrls = useMemo(() => {
    const configured = items.filter((item) => item.url?.trim())
    if (configured.length > 0) return configured
    if (!fallbackBase) return []
    return [
      {
        url: fallbackBase,
        route: t('Direct'),
        description: t('Direct connection'),
        color: 'blue',
      } satisfies ApiInfoItem,
    ]
  }, [fallbackBase, items, t])

  const sampleBase = accessUrls[0]?.url
    ? trimTrailingSlash(accessUrls[0].url)
    : fallbackBase || 'https://your-api.example.com'

  return (
    <PublicLayout showMainContainer={false}>
      <div className='pt-14'>
        <div className='mx-auto flex max-w-3xl flex-col gap-10 px-4 py-10 sm:px-6'>
          <header className='space-y-2'>
            <h1 className='text-2xl font-semibold tracking-tight'>
              {t('API integration')}
            </h1>
            <p className='text-muted-foreground text-sm leading-relaxed'>
              {t(
                'Use one API key with OpenAI-compatible endpoints for chat, images, video, and more.'
              )}
            </p>
          </header>

          {isLoading ? (
            <div className='flex flex-col gap-3'>
              <Skeleton className='h-8 w-[45%]' />
              <Skeleton className='h-4 w-full' />
              <Skeleton className='h-4 w-[90%]' />
            </div>
          ) : null}

          {!isLoading && hasCustomContent ? (
            <SiteDocument content={customContent} variant='main' />
          ) : null}

          {!isLoading && !hasCustomContent ? (
            <>
          <section className='space-y-3'>
            <h2 className='text-lg font-medium'>{t('Get started')}</h2>
            <ol className='text-muted-foreground list-decimal space-y-2 pl-5 text-sm leading-relaxed'>
              <li>
                {t('Create an API key in')}{' '}
                {isAuthenticated ? (
                  <Link to='/keys' className='text-foreground underline-offset-4 hover:underline'>
                    {t('API Keys')}
                  </Link>
                ) : (
                  <Link
                    to='/sign-in'
                    className='text-foreground underline-offset-4 hover:underline'
                  >
                    {t('Sign in')}
                  </Link>
                )}
                .
              </li>
              <li>{t('Pick a base URL. Use the Cloudflare URL in mainland China.')}</li>
              <li>
                {t('Send requests with')}{' '}
                <code className='bg-muted rounded px-1 py-0.5 font-mono text-xs'>
                  Authorization: Bearer $API_KEY
                </code>
                .
              </li>
            </ol>
            {accessUrls.length > 0 && (
              <ul className='grid gap-2'>
                {accessUrls.map((item) => (
                  <li
                    key={`${item.route}-${item.url}`}
                    className='bg-muted/30 flex min-w-0 items-center gap-2.5 rounded-lg border px-3 py-2.5'
                  >
                    <span
                      className={`size-2 shrink-0 rounded-full ${getBgColorClass(item.color)}`}
                    />
                    <div className='min-w-0 flex-1'>
                      <p className='text-sm font-medium'>{t(item.route)}</p>
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
            )}
          </section>

          <section className='space-y-3'>
            <h2 className='text-lg font-medium'>{t('API protocols')}</h2>
            <div className='space-y-3'>
              {PROTOCOL_GUIDES.map((guide) => {
                const sample = curlSample(sampleBase, guide.path, guide.body)
                return (
                  <article
                    key={guide.id}
                    className='space-y-2 rounded-lg border px-3 py-3'
                  >
                    <div className='flex flex-wrap items-baseline justify-between gap-2'>
                      <h3 className='text-sm font-medium'>
                        {t(guide.titleKey)}
                      </h3>
                      <code className='text-muted-foreground font-mono text-xs'>
                        POST {guide.path}
                      </code>
                    </div>
                    <div className='relative'>
                      <pre className='bg-muted/40 overflow-x-auto rounded-md p-3 pr-10 font-mono text-xs leading-relaxed'>
                        {sample}
                      </pre>
                      <CopyButton
                        value={sample}
                        variant='ghost'
                        size='sm'
                        className='absolute top-1.5 right-1.5 size-7 p-0'
                        iconClassName='size-3.5'
                        tooltip={t('Copy to clipboard')}
                      />
                    </div>
                  </article>
                )
              })}
            </div>
          </section>

          <section className='space-y-3'>
            <h2 className='text-lg font-medium'>{t('Next steps')}</h2>
            <ul className='text-muted-foreground space-y-2 text-sm'>
              <li>
                <Link
                  to={isAuthenticated ? '/playground' : '/sign-in'}
                  className='text-foreground underline-offset-4 hover:underline'
                >
                  {t('Playground')}
                </Link>
                {' — '}
                {t('Try models in the browser before writing client code.')}
              </li>
              <li>
                <Link
                  to='/pricing'
                  className='text-foreground underline-offset-4 hover:underline'
                >
                  {t('Model Square')}
                </Link>
                {' — '}
                {t('See model prices and per-model request examples.')}
              </li>
            </ul>
          </section>
            </>
          ) : null}
        </div>
        <Footer />
      </div>
    </PublicLayout>
  )
}
