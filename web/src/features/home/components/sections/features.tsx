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
import {
  BarChart3,
  Code,
  Image,
  KeyRound,
  MessageSquare,
  Video,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'

interface FeaturesProps {
  className?: string
}

export function Features(_props: FeaturesProps) {
  const { t } = useTranslation()

  const features = [
    {
      id: 'models',
      num: '01',
      title: t('One key, many models'),
      desc: t(
        'Call OpenAI, Claude, Gemini, DeepSeek, and Qwen through a single token.'
      ),
      span: 'md:col-span-2',
      visual: (
        <div className='mt-4 grid grid-cols-3 gap-2'>
          {['OpenAI', 'Claude', 'Gemini', 'DeepSeek', 'Qwen', 'Llama'].map(
            (name) => (
              <div
                key={name}
                className='border-border bg-muted/30 text-muted-foreground hover:text-foreground flex items-center justify-center rounded-lg border px-3 py-2 text-xs transition-colors duration-200'
              >
                {name}
              </div>
            )
          )}
        </div>
      ),
    },
    {
      id: 'modalities',
      num: '02',
      title: t('Chat, code, image, and video'),
      desc: t(
        'OpenAI-compatible routes for completions, images, and video tasks.'
      ),
      span: 'md:col-span-1',
      visual: (
        <div className='mt-4 grid grid-cols-2 gap-2'>
          {[
            { icon: MessageSquare, label: t('Chat') },
            { icon: Image, label: t('Image') },
            { icon: Video, label: t('Video') },
            { icon: Code, label: t('Code') },
          ].map((item) => {
            const Icon = item.icon
            return (
              <div
                key={item.label}
                className='border-border bg-muted/30 text-muted-foreground flex items-center gap-1.5 rounded-lg border px-2.5 py-2 text-xs'
              >
                <Icon className='size-3.5 shrink-0' strokeWidth={1.75} />
                {item.label}
              </div>
            )
          })}
        </div>
      ),
    },
    {
      id: 'billing',
      num: '03',
      title: t('Usage and billing in one place'),
      desc: t(
        'See cost and traffic for every chat, image, and video request.'
      ),
      span: 'md:col-span-1',
      visual: (
        <div className='mt-4 space-y-2'>
          {[t('Chat'), t('Image Generation'), t('Video')].map((step, i) => (
            <div key={step} className='flex items-center gap-2'>
              <div className='border-border bg-muted text-muted-foreground flex size-6 items-center justify-center rounded-full border text-[10px] font-bold'>
                {i + 1}
              </div>
              <div className='bg-border h-px flex-1' />
              <span className='text-muted-foreground text-xs'>{step}</span>
            </div>
          ))}
        </div>
      ),
    },
    {
      id: 'compatible',
      num: '04',
      title: t('Drop into existing clients'),
      desc: t(
        'Point your OpenAI-compatible app at this gateway and keep your current SDK.'
      ),
      span: 'md:col-span-2',
      visual: (
        <div className='text-muted-foreground mt-4 font-mono text-xs'>
          <div>OPENAI_BASE_URL=https://your-gateway/v1</div>
          <div>OPENAI_API_KEY=sk-••••</div>
        </div>
      ),
    },
  ]

  const steps = [
    {
      num: '1',
      title: t('Get a key'),
      desc: t('Create a token after you sign in.'),
      icon: <KeyRound className='size-5' strokeWidth={1.5} />,
    },
    {
      num: '2',
      title: t('Call the gateway'),
      desc: t('Send chat, image, and video requests with the same base URL.'),
      icon: <MessageSquare className='size-5' strokeWidth={1.5} />,
    },
    {
      num: '3',
      title: t('Monitor'),
      desc: t('Track usage, costs and performance with real-time analytics'),
      icon: <BarChart3 className='size-5' strokeWidth={1.5} />,
    },
  ]

  return (
    <section className='relative z-10 px-6 py-24 md:py-32'>
      <div className='mx-auto max-w-6xl'>
        <AnimateInView className='mb-10 max-w-lg md:mb-12'>
          <p className='text-muted-foreground mb-3 text-xs font-medium'>
            {t('How it works')}
          </p>
          <h2 className='text-2xl leading-tight font-semibold tracking-tight md:text-3xl'>
            {t('One gateway for the models your apps already use')}
          </h2>
        </AnimateInView>

        <div className='mb-10 grid gap-6 md:mb-12 md:grid-cols-3 md:gap-8'>
          {steps.map((step, i) => (
            <AnimateInView
              key={step.num}
              delay={i * 80}
              animation='fade-up'
              className='flex gap-3'
            >
              <div className='text-muted-foreground border-border bg-muted/30 flex size-10 shrink-0 items-center justify-center rounded-lg border'>
                {step.icon}
              </div>
              <div className='min-w-0'>
                <h3 className='text-sm font-semibold'>
                  {step.num}. {step.title}
                </h3>
                <p className='text-muted-foreground mt-1 text-sm leading-relaxed'>
                  {step.desc}
                </p>
              </div>
            </AnimateInView>
          ))}
        </div>

        <div className='border-border bg-border grid gap-px overflow-hidden rounded-lg border md:grid-cols-3'>
          {features.map((f, i) => (
            <AnimateInView
              key={f.id}
              delay={i * 100}
              animation='scale-in'
              className={`bg-background group hover:bg-muted/20 p-7 transition-colors duration-300 md:p-8 ${f.span}`}
            >
              <div className='mb-3 flex items-center gap-3'>
                <span className='border-border/40 bg-muted text-muted-foreground flex size-7 items-center justify-center rounded-md border text-[10px] font-semibold tabular-nums'>
                  {f.num}
                </span>
                <h3 className='text-sm font-semibold'>{f.title}</h3>
              </div>
              <p className='text-muted-foreground text-sm leading-relaxed'>
                {f.desc}
              </p>
              {f.visual}
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
