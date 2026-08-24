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
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

import { formatSquareGroupRatio, type SquarePricePair } from '../lib/price'

type SquarePricePairViewProps = SquarePricePair & {
  className?: string
  align?: 'start' | 'end'
}

export function SquarePricePairView(props: SquarePricePairViewProps) {
  const { t } = useTranslation()
  const ratioLabel = formatSquareGroupRatio(props.groupRatio)
  const title = props.differs
    ? `${t('Model price')}: ${props.own}; ${t('After group ratio')}: ${props.grouped}`
    : `${t('Model price')}: ${props.own}`

  if (!props.differs) {
    return (
      <span
        className={cn(
          'text-foreground font-mono font-semibold',
          props.className
        )}
        title={title}
      >
        {props.own}
      </span>
    )
  }

  return (
    <span
      className={cn(
        'inline-flex min-w-0 flex-col',
        props.align === 'end' ? 'items-end' : 'items-start',
        props.className
      )}
      title={title}
    >
      <span className='text-foreground font-mono font-semibold'>
        {props.own}
      </span>
      <span className='text-muted-foreground font-mono text-[11px] leading-tight'>
        {props.grouped}
        <span className='ml-1 opacity-70'>×{ratioLabel}</span>
      </span>
    </span>
  )
}
