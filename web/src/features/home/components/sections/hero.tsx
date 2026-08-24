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
import { Link } from '@tanstack/react-router'
import { ArrowRight, Code, Image, MessageSquare, Video } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { useStatus } from '@/hooks/use-status'

import { HeroTerminalDemo } from '../hero-terminal-demo'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

const WORKLOADS = [
  { id: 'chat', labelKey: 'Chat', icon: MessageSquare },
  { id: 'code', labelKey: 'Code', icon: Code },
  { id: 'image', labelKey: 'Image', icon: Image },
  { id: 'video', labelKey: 'Video', icon: Video },
] as const

export function Hero(props: HeroProps) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const registerEnabled =
    !status?.self_use_mode_enabled && status?.register_enabled !== false

  return (
    <section className='relative z-10 overflow-hidden px-6 pt-8 pb-16 md:pt-12 md:pb-24 lg:pt-16 lg:pb-28'>
      <div className='mx-auto grid max-w-6xl grid-cols-1 items-start gap-12 lg:grid-cols-12 lg:gap-8'>
        {/* Left Column: Title, description, action buttons and application support */}
        <div className='flex flex-col items-start text-left lg:col-span-6'>
          {/* Top Pill Badge */}
          <div
            className='landing-animate-fade-up border-border bg-muted text-muted-foreground mb-5 inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-[11px] font-medium opacity-0 shadow-xs'
            style={{ animationDelay: '0ms' }}
          >
            <span className='relative flex size-1.5'>
              <span className='bg-foreground/30 absolute inline-flex h-full w-full animate-ping rounded-full opacity-75' />
              <span className='bg-foreground/70 relative inline-flex size-1.5 rounded-full' />
            </span>
            <span>{t('API relay for mainstream models')}</span>
          </div>

          <h1
            className='landing-animate-fade-up text-[clamp(1.875rem,3.6vw,2.5rem)] leading-[1.2] font-semibold tracking-tight'
            style={{ animationDelay: '60ms' }}
          >
            {t('One key for chat, code, image, and video')}
          </h1>
          <p
            className='landing-animate-fade-up text-muted-foreground/80 mt-5 max-w-xl text-base leading-relaxed opacity-0 md:text-[15px]'
            style={{ animationDelay: '120ms' }}
          >
            {t(
              'Route OpenAI, Claude, Gemini, and more through OpenAI-compatible APIs. Use the same token for chat, coding, image generation, and video.'
            )}
          </p>

          {!props.isAuthenticated && (
            <div
              className='landing-animate-fade-up mt-8 flex flex-wrap items-center gap-3 opacity-0'
              style={{ animationDelay: '180ms' }}
            >
              <Button
                className='group h-11 rounded-lg px-5 text-sm font-medium'
                render={<Link to='/sign-in' />}
              >
                {t('Sign in')}
                <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
              </Button>
              {registerEnabled && (
                <Button
                  variant='outline'
                  className='border-border/50 hover:border-border hover:bg-muted/50 h-11 rounded-lg px-5 text-sm font-medium'
                  render={<Link to='/sign-up' />}
                >
                  {t('Sign up')}
                </Button>
              )}
            </div>
          )}

          <div
            className='landing-animate-fade-up mt-10 w-full max-w-xl opacity-0'
            style={{ animationDelay: '240ms' }}
          >
            <div className='mb-4 flex flex-col gap-1'>
              <span className='text-muted-foreground/50 text-[10px] font-bold tracking-[0.15em] uppercase'>
                {t('What you can call')}
              </span>
              <p className='text-muted-foreground/60 text-xs leading-relaxed'>
                {t(
                  'Use one API key for chat, coding, image generation, and video tasks.'
                )}
              </p>
            </div>
            <div className='flex flex-wrap items-center gap-2'>
              {WORKLOADS.map((item) => {
                const Icon = item.icon
                return (
                  <div
                    key={item.id}
                    className='border-border/40 bg-muted/15 text-foreground/80 flex items-center gap-2 rounded-full border px-3.5 py-2 text-sm font-medium'
                  >
                    <Icon className='size-3.5 shrink-0' strokeWidth={1.75} />
                    <span>{t(item.labelKey)}</span>
                  </div>
                )
              })}
            </div>
          </div>
        </div>

        {/* Right Column: Hero Terminal API Demo */}
        <div
          className='landing-animate-fade-up flex w-full justify-center opacity-0 lg:col-span-6'
          style={{ animationDelay: '320ms' }}
        >
          <HeroTerminalDemo className='mt-8 lg:mt-0' />
        </div>
      </div>
    </section>
  )
}
