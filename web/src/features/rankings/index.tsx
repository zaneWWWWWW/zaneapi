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
import { useNavigate, useSearch } from '@tanstack/react-router'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { ModulePageFrame } from '@/components/layout'
import { Skeleton } from '@/components/ui/skeleton'

import {
  MarketShareSection,
  ModelsSection,
  PulseSection,
  RankingsHero,
} from './components'
import { useRankings } from './hooks/use-rankings'
import type { RankingPeriod } from './types'

const VALID_PERIODS: RankingPeriod[] = ['today', 'week', 'month', 'year']

export function Rankings() {
  const { t } = useTranslation()
  const search = useSearch({ from: '/_authenticated/rankings/' })
  const navigate = useNavigate()

  const period: RankingPeriod = VALID_PERIODS.includes(
    search.period as RankingPeriod
  )
    ? (search.period as RankingPeriod)
    : 'week'

  const rankingsQuery = useRankings(period)
  const snapshot = rankingsQuery.data?.data

  const handlePeriodChange = (next: RankingPeriod) => {
    navigate({
      to: '/rankings',
      search: (prev) => ({ ...prev, period: next }),
    })
  }

  let rankingsBody: ReactNode
  if (rankingsQuery.isLoading) {
    rankingsBody = <RankingsLoading />
  } else if (!snapshot) {
    const errorMessage =
      rankingsQuery.error instanceof Error
        ? rankingsQuery.error.message
        : t('Unable to load rankings data')
    rankingsBody = <RankingsError message={errorMessage} />
  } else {
    rankingsBody = (
      <>
        <ModelsSection
          history={snapshot.models_history}
          rows={snapshot.models}
          period={period}
        />

        <MarketShareSection
          history={snapshot.vendor_share_history}
          rows={snapshot.vendors}
          period={period}
        />

        <PulseSection
          movers={snapshot.top_movers}
          droppers={snapshot.top_droppers}
        />
      </>
    )
  }

  return (
    <ModulePageFrame contentClassName='max-w-[1280px] space-y-8'>
      <RankingsHero period={period} onPeriodChange={handlePeriodChange} />
      {rankingsBody}
    </ModulePageFrame>
  )
}

function RankingsLoading() {
  return (
    <div className='space-y-6'>
      <Skeleton className='h-[420px] w-full rounded-lg' />
      <Skeleton className='h-[360px] w-full rounded-lg' />
      <Skeleton className='h-[180px] w-full rounded-lg' />
    </div>
  )
}

function RankingsError(props: { message: string }) {
  const { t } = useTranslation()
  return (
    <div className='bg-card rounded-lg border border-dashed px-6 py-12 text-center'>
      <h2 className='text-foreground text-base font-semibold'>
        {t('Unable to load rankings')}
      </h2>
      <p className='text-muted-foreground mx-auto mt-2 max-w-md text-sm'>
        {props.message}
      </p>
    </div>
  )
}
