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
import {
  ArrowRight,
  Check,
  ChevronDown,
  ChevronUp,
  Circle,
  CreditCard,
  KeyRound,
  ListChecks,
  TerminalSquare,
  type LucideIcon,
} from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import {
  CardStaggerContainer,
  CardStaggerItem,
} from '@/components/page-transition'
import { Button } from '@/components/ui/button'
import { getApiKeys } from '@/features/keys/api'
import type { ApiKey } from '@/features/keys/types'
import { ROLE } from '@/lib/roles'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

import { useDashboardContentVisibility } from '../../hooks/use-status-data'
import { AnnouncementsPanel } from './announcements-panel'
import { ApiInfoPanel } from './api-info-panel'
import { FAQPanel } from './faq-panel'
import { GroupAvailabilityPanel } from './group-availability-panel'
import { getOverviewContentLayout } from './overview-content-layout'
import { PerformanceHealthPanel } from './performance-health-panel'
import { SummaryCards } from './summary-cards'

const SETUP_GUIDE_VISIBILITY_STORAGE_KEY =
  'dashboard_overview_setup_guide_expanded'

type DashboardActionPath = '/keys' | '/wallet' | '/playground'

interface StartStep {
  title: string
  description: string
  to: DashboardActionPath
  icon: LucideIcon
  completed: boolean
}

function getSavedSetupGuideExpanded(): boolean | null {
  if (typeof window === 'undefined') return null
  const saved = window.localStorage.getItem(SETUP_GUIDE_VISIBILITY_STORAGE_KEY)
  if (saved === 'expanded') return true
  if (saved === 'collapsed') return false
  return null
}

function saveSetupGuideExpanded(expanded: boolean): void {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(
    SETUP_GUIDE_VISIBILITY_STORAGE_KEY,
    expanded ? 'expanded' : 'collapsed'
  )
}

function getPreferredKey(keys: ApiKey[]): ApiKey | null {
  return keys.find((item) => item.status === 1) ?? keys[0] ?? null
}

function StartStepItem(props: {
  step: StartStep
  index: number
  isLast: boolean
}) {
  const Icon = props.step.icon
  const StatusIcon = props.step.completed ? Check : Circle

  return (
    <li className='relative flex gap-3 pb-2.5 last:pb-0'>
      {!props.isLast && (
        <span
          className='bg-border absolute top-9 bottom-0 left-4 w-px'
          aria-hidden='true'
        />
      )}
      <span
        className={cn(
          'bg-background relative z-10 flex size-8 shrink-0 items-center justify-center rounded-lg border shadow-xs',
          props.step.completed && 'border-success/30 bg-success/10'
        )}
      >
        <StatusIcon
          className={props.step.completed ? 'text-success size-4' : 'size-4'}
          aria-hidden='true'
        />
      </span>

      <Link
        to={props.step.to}
        className='bg-background/70 hover:bg-muted/50 focus-visible:ring-ring flex min-w-0 flex-1 items-center justify-between gap-3 rounded-lg border px-3 py-2.5 text-left shadow-xs transition-colors outline-none focus-visible:ring-2'
      >
        <span className='flex min-w-0 items-start gap-2.5'>
          <span className='bg-muted mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg'>
            <Icon className='size-3.5' aria-hidden='true' />
          </span>
          <span className='flex min-w-0 flex-col gap-0.5'>
            <span className='flex items-center gap-2 text-sm font-medium'>
              <span className='text-muted-foreground font-mono text-xs tabular-nums'>
                {props.index + 1}.
              </span>
              <span className='truncate'>{props.step.title}</span>
            </span>
            <span className='text-muted-foreground line-clamp-1 text-xs'>
              {props.step.description}
            </span>
          </span>
        </span>
        <ArrowRight
          className='text-muted-foreground size-4 shrink-0'
          aria-hidden='true'
        />
      </Link>
    </li>
  )
}

export function OverviewDashboard() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  const {
    apiInfo: showApiInfoPanel,
    announcements: showAnnouncementsPanel,
    faq: showFAQPanel,
    groupAvailability: showGroupAvailability,
  } = useDashboardContentVisibility()
  const [manualSetupGuideExpanded, setManualSetupGuideExpanded] = useState<
    boolean | null
  >(() => getSavedSetupGuideExpanded())

  const requestCount = Number(user?.request_count ?? 0)
  const remainQuota = Number(user?.quota ?? 0)
  const usedQuota = Number(user?.used_quota ?? 0)
  const isAdmin = Boolean(user?.role && user.role >= ROLE.ADMIN)

  const apiKeysQuery = useQuery({
    queryKey: ['dashboard', 'overview', 'api-keys'],
    queryFn: async () => {
      const result = await getApiKeys({ p: 1, size: 10 })
      return result.success ? (result.data?.items ?? []) : []
    },
    staleTime: 60 * 1000,
  })

  const preferredKey = useMemo(
    () => getPreferredKey(apiKeysQuery.data ?? []),
    [apiKeysQuery.data]
  )

  const startSteps = useMemo<StartStep[]>(
    () => [
      {
        title: t('Create API Key'),
        description: t('Create a key for your app or service'),
        to: '/keys',
        icon: KeyRound,
        completed: Boolean(preferredKey),
      },
      {
        title: t('Add credits'),
        description: t('Keep enough balance before production traffic'),
        to: '/wallet',
        icon: CreditCard,
        completed: remainQuota > 0 || usedQuota > 0,
      },
      {
        title: t('Send a request'),
        description: t('Verify routing with Playground or your client'),
        to: '/playground',
        icon: TerminalSquare,
        completed: requestCount > 0,
      },
    ],
    [preferredKey, remainQuota, requestCount, t, usedQuota]
  )

  const completedStepCount = startSteps.filter((step) => step.completed).length
  const setupComplete = completedStepCount === startSteps.length
  const setupStatusReady = apiKeysQuery.isFetched && Boolean(user)
  const setupGuideExpanded =
    manualSetupGuideExpanded ?? (setupStatusReady && !setupComplete)
  const contentLayout = getOverviewContentLayout({
    isAdmin,
    showApiInfo: showApiInfoPanel,
    showAnnouncements: showAnnouncementsPanel,
    showFaq: showFAQPanel,
    showGroupAvailability,
  })

  const handleSetupGuideToggle = () => {
    const nextExpanded = !setupGuideExpanded
    setManualSetupGuideExpanded(nextExpanded)
    saveSetupGuideExpanded(nextExpanded)
  }

  return (
    <div className='flex flex-col gap-4'>
      {setupGuideExpanded ? (
        <CardStaggerContainer>
          <CardStaggerItem className='bg-card overflow-hidden rounded-lg border shadow-xs'>
            <div className='p-4 sm:p-5'>
              <div className='mb-4 flex flex-wrap items-start justify-between gap-3'>
                <div className='flex min-w-0 flex-col gap-1'>
                  <div className='text-muted-foreground flex items-center gap-2 text-xs font-medium'>
                    <ListChecks className='size-3.5' aria-hidden='true' />
                    {t('Get started')}
                  </div>
                  <h3 className='text-lg font-semibold tracking-tight'>
                    {t('Build on your API gateway in minutes')}
                  </h3>
                </div>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={handleSetupGuideToggle}
                >
                  <ChevronUp data-icon='inline-start' />
                  {t('Hide setup guide')}
                </Button>
              </div>

              <ol>
                {startSteps.map((step, index) => (
                  <StartStepItem
                    key={step.title}
                    step={step}
                    index={index}
                    isLast={index === startSteps.length - 1}
                  />
                ))}
              </ol>
            </div>
          </CardStaggerItem>
        </CardStaggerContainer>
      ) : (
        <CardStaggerContainer>
          <CardStaggerItem className='bg-card overflow-hidden rounded-lg border shadow-xs'>
            <div className='flex flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-5'>
              <div className='flex min-w-0 items-center gap-3'>
                <span className='bg-muted flex size-8 shrink-0 items-center justify-center rounded-lg'>
                  <Check className='text-success size-4' aria-hidden='true' />
                </span>
                <div className='min-w-0'>
                  <h3 className='truncate text-sm font-semibold'>
                    {setupComplete
                      ? t('Setup guide complete')
                      : t('Setup guide')}
                  </h3>
                  <p className='text-muted-foreground text-xs'>
                    {t('Setup progress: {{completed}}/{{total}}', {
                      completed: completedStepCount,
                      total: startSteps.length,
                    })}
                  </p>
                </div>
              </div>
              <Button
                variant='outline'
                size='sm'
                onClick={handleSetupGuideToggle}
              >
                <ChevronDown data-icon='inline-start' />
                {t('Show setup guide')}
              </Button>
            </div>
          </CardStaggerItem>
        </CardStaggerContainer>
      )}

      <SummaryCards />

      {contentLayout.show && (
        <CardStaggerContainer className={contentLayout.outerClassName}>
          {contentLayout.showLeft && (
            <CardStaggerItem className={contentLayout.leftClassName}>
              {contentLayout.showPerformance && <PerformanceHealthPanel />}
              {contentLayout.pairApiAndAnnouncements ? (
                <div className={contentLayout.apiAnnouncementsRowClassName}>
                  <AnnouncementsPanel />
                  <ApiInfoPanel />
                </div>
              ) : (
                <>
                  {contentLayout.showAnnouncements && <AnnouncementsPanel />}
                  {contentLayout.showApiInfo && <ApiInfoPanel />}
                </>
              )}
              {contentLayout.showFaq && <FAQPanel />}
            </CardStaggerItem>
          )}
          {contentLayout.showGroupAvailability && (
            <CardStaggerItem
              className={contentLayout.groupAvailabilityItemClassName}
            >
              <GroupAvailabilityPanel
                fill={contentLayout.groupAvailabilityFill}
              />
            </CardStaggerItem>
          )}
        </CardStaggerContainer>
      )}
    </div>
  )
}
