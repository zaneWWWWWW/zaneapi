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
  Box,
  CreditCard,
  FileText,
  FlaskConical,
  Key,
  LayoutDashboard,
  MessageSquare,
  Radio,
  Store,
  Ticket,
  Trophy,
  Users,
  Wallet,
} from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { getSystemSettingsNavItems } from '@/components/layout/config/system-settings.config'
import type { NavItem, SidebarData } from '@/components/layout/types'
import { useStatus } from '@/hooks/use-status'
import { parseHeaderNavModulesFromStatus } from '@/lib/nav-modules'

/**
 * Root navigation groups for the application sidebar.
 *
 * System settings sections are listed under Admin instead of replacing
 * this tree with a drill-in view.
 */
export function useSidebarData(): SidebarData {
  const { t } = useTranslation()
  const { status } = useStatus()

  return useMemo(() => {
    const modules = parseHeaderNavModulesFromStatus(
      status as Record<string, unknown> | null
    )
    const exploreItems: NavItem[] = []

    if (
      modules.pricing &&
      typeof modules.pricing === 'object' &&
      modules.pricing.enabled
    ) {
      exploreItems.push({
        title: t('Model Square'),
        url: '/pricing',
        icon: Store,
      })
    }

    if (
      modules.rankings &&
      typeof modules.rankings === 'object' &&
      modules.rankings.enabled
    ) {
      exploreItems.push({
        title: t('Rankings'),
        url: '/rankings',
        icon: Trophy,
      })
    }

    const mainItems: NavItem[] = [
      ...exploreItems,
      {
        title: t('Playground'),
        url: '/playground',
        icon: FlaskConical,
      },
      {
        title: t('Chat'),
        icon: MessageSquare,
        type: 'chat-presets',
      },
      {
        title: t('Dashboard'),
        url: '/dashboard/overview',
        icon: LayoutDashboard,
        activeUrls: [
          '/dashboard/models',
          '/dashboard/flow',
          '/dashboard/users',
        ],
      },
      {
        title: t('API Keys'),
        url: '/keys',
        icon: Key,
      },
      {
        title: t('Usage Logs'),
        url: '/usage-logs/common',
        icon: FileText,
        activeUrls: ['/usage-logs/drawing', '/usage-logs/task'],
        configUrls: [
          '/usage-logs/common',
          '/usage-logs/drawing',
          '/usage-logs/task',
        ],
      },
      {
        title: t('Wallet'),
        url: '/wallet',
        icon: Wallet,
      },
    ]

    const navGroups: SidebarData['navGroups'] = []
    if (mainItems.length > 0) {
      navGroups.push({
        id: 'main',
        items: mainItems,
      })
    }

    navGroups.push({
      id: 'admin',
      title: t('Admin'),
      items: [
        {
          title: t('Channels'),
          url: '/channels',
          icon: Radio,
        },
        {
          title: t('Models'),
          url: '/models/metadata',
          icon: Box,
          activeUrls: ['/models/deployments'],
        },
        {
          title: t('Users'),
          url: '/users',
          icon: Users,
        },
        {
          title: t('Redemption Codes'),
          url: '/redemption-codes',
          icon: Ticket,
        },
        {
          title: t('Subscriptions'),
          url: '/subscriptions',
          icon: CreditCard,
        },
      ],
    })

    navGroups.push({
      id: 'system-settings',
      title: t('System Settings'),
      items: getSystemSettingsNavItems(t),
    })

    return { navGroups }
  }, [status, t])
}
