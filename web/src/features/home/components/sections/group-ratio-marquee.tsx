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
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { useStatus } from '@/hooks/use-status'

import type { HomePageGroupRatio } from '../../types'

function formatGroupRatio(ratio: number): string {
  if (!Number.isFinite(ratio)) return '-'
  const text = ratio % 1 === 0 ? String(ratio) : ratio.toFixed(4).replace(/\.?0+$/, '')
  return `${text}x`
}

function parseGroupRatios(value: unknown): HomePageGroupRatio[] {
  if (!Array.isArray(value)) return []
  const items: HomePageGroupRatio[] = []
  for (const entry of value) {
    if (!entry || typeof entry !== 'object') continue
    const record = entry as Record<string, unknown>
    if (typeof record.name !== 'string' || record.name === '') continue
    const ratio =
      typeof record.ratio === 'number' ? record.ratio : Number(record.ratio)
    items.push({
      name: record.name,
      description:
        typeof record.description === 'string' && record.description !== ''
          ? record.description
          : record.name,
      ratio: Number.isFinite(ratio) ? ratio : 1,
    })
  }
  return items
}

function GroupRatioCard(props: { item: HomePageGroupRatio }) {
  return (
    <div className='border-border bg-background flex min-w-[168px] shrink-0 items-center gap-3 rounded-lg border px-3.5 py-2'>
      <div className='min-w-0 flex-1'>
        <span className='block truncate text-sm font-semibold'>
          {props.item.description}
        </span>
        <span className='text-muted-foreground block truncate font-mono text-[11px]'>
          {props.item.name}
        </span>
      </div>
      <span className='text-base font-semibold tabular-nums'>
        {formatGroupRatio(props.item.ratio)}
      </span>
    </div>
  )
}

export function GroupRatioMarquee() {
  const { t } = useTranslation()
  const { status } = useStatus()
  const items = parseGroupRatios(status?.home_page_group_ratios)
  const [reduceMotion, setReduceMotion] = useState(false)

  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
    setReduceMotion(mq.matches)
    const onChange = () => setReduceMotion(mq.matches)
    mq.addEventListener('change', onChange)
    return () => mq.removeEventListener('change', onChange)
  }, [])

  const trackItems = useMemo(() => {
    if (items.length === 0) return []
    const repeats = Math.max(1, Math.ceil(6 / items.length))
    const unit = Array.from({ length: repeats }, (_, copyIndex) =>
      items.map((item) => ({
        ...item,
        key: `${copyIndex}-${item.name}`,
      }))
    ).flat()
    return [
      ...unit.map((item) => ({ ...item, key: `a-${item.key}` })),
      ...unit.map((item) => ({ ...item, key: `b-${item.key}` })),
    ]
  }, [items])

  if (items.length === 0) {
    return null
  }

  const durationSec = Math.max(20, items.length * 8)

  return (
    <section
      className='border-border/40 bg-muted/10 relative z-10 border-b px-6 py-3 md:py-3.5'
      aria-label={t('Site group ratios')}
    >
      <div className='mx-auto max-w-6xl'>
        <p className='text-muted-foreground mb-2 text-[11px] font-medium tracking-wide uppercase'>
          {t('Site group ratios')}
        </p>
        <ul className='sr-only'>
          {items.map((item) => (
            <li key={item.name}>
              {item.description}: {formatGroupRatio(item.ratio)}
            </li>
          ))}
        </ul>
        {reduceMotion ? (
          <div className='flex flex-wrap gap-3' aria-hidden='true'>
            {items.map((item) => (
              <GroupRatioCard key={item.name} item={item} />
            ))}
          </div>
        ) : (
          <div className='homepage-group-marquee-viewport' aria-hidden='true'>
            <div
              className='homepage-group-marquee flex w-max gap-3'
              style={{ animationDuration: `${durationSec}s` }}
            >
              {trackItems.map((item) => (
                <GroupRatioCard key={item.key} item={item} />
              ))}
            </div>
          </div>
        )}
      </div>
    </section>
  )
}
