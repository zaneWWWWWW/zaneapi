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
import type { TFunction } from 'i18next'
import {
  CreditCard,
  GitBranch,
  Layout,
  Settings,
  Shield,
  ShieldAlert,
  Wrench,
} from 'lucide-react'

import { getAuthSectionNavItems } from '@/features/system-settings/auth/section-registry.tsx'
import { getBillingSectionNavItems } from '@/features/system-settings/billing/section-registry.tsx'
import { getContentSectionNavItems } from '@/features/system-settings/content/section-registry.tsx'
import { getModelsSectionNavItems } from '@/features/system-settings/models/section-registry.tsx'
import { getOperationsSectionNavItems } from '@/features/system-settings/operations/section-registry.tsx'
import { getSecuritySectionNavItems } from '@/features/system-settings/security/section-registry.tsx'
import { getSiteSectionNavItems } from '@/features/system-settings/site/section-registry.tsx'
import { ROLE } from '@/lib/roles'

import type { NavGroup, NavItem, SidebarView } from '../types'

/**
 * System settings sections shown in their own sidebar group.
 */
export function getSystemSettingsNavItems(t: TFunction): NavItem[] {
  return [
    {
      title: t('Site & Branding'),
      icon: Settings,
      items: getSiteSectionNavItems(t),
    },
    {
      title: t('Authentication'),
      icon: Shield,
      items: getAuthSectionNavItems(t),
    },
    {
      title: t('Billing & Payment'),
      icon: CreditCard,
      items: getBillingSectionNavItems(t),
    },
    {
      title: t('Routing & Reliability'),
      icon: GitBranch,
      items: getModelsSectionNavItems(t),
    },
    {
      title: t('Security & Limits'),
      icon: ShieldAlert,
      items: getSecuritySectionNavItems(t),
    },
    {
      title: t('Console display'),
      icon: Layout,
      url: '/system-settings/content/dashboard',
      activeUrls: getContentSectionNavItems(t).map((item) => item.url),
    },
    {
      title: t('Operations'),
      icon: Wrench,
      items: [
        {
          title: t('Runtime'),
          url: '/system-info',
          requiredRole: ROLE.SUPER_ADMIN,
        },
        ...getOperationsSectionNavItems(t),
      ],
    },
  ]
}

function getSystemSettingsNavGroups(t: TFunction): NavGroup[] {
  return [
    {
      id: 'system-administration',
      title: t('System Administration'),
      items: getSystemSettingsNavItems(t),
    },
  ]
}

/**
 * Nested sidebar view for `/system-settings/*`.
 *
 * Activates the Vercel / Cloudflare-style drill-in sidebar:
 * the root navigation is replaced by the system administration
 * groups, with a "Back to Dashboard" affordance in the header.
 */
export const SYSTEM_SETTINGS_VIEW: SidebarView = {
  id: 'system-settings',
  pathPattern: /^\/system-settings(\/|$)/,
  parent: {
    to: '/dashboard/overview',
    label: 'Back to Dashboard',
  },
  getNavGroups: getSystemSettingsNavGroups,
}
