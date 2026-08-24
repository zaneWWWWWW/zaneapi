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
import { SettingsPage } from '../components/settings-page'
import type { SiteSettings, SystemOption } from '../types'
import {
  SITE_DEFAULT_SECTION,
  getSiteSectionContent,
  getSiteSectionMeta,
} from './section-registry.tsx'

const defaultSiteSettings: SiteSettings = {
  Notice: '',
  SystemName: 'Macroapple',
  Logo: '',
  Footer: '',
  About: '',
  HomePageContent: '',
  HomePageDisplayedGroups: '[]',
  GroupRatio: '{}',
  UserUsableGroups: '{}',
  ServerAddress: '',
  'legal.user_agreement': '',
  'legal.privacy_policy': '',
  HeaderNavModules: '',
  SidebarModulesAdmin: '',
  'console_setting.announcements': '[]',
  'console_setting.announcements_enabled': true,
  'console_setting.faq': '[]',
  'console_setting.faq_enabled': true,
  'console_setting.api_info': '[]',
  'console_setting.api_info_enabled': true,
}

function resolveSiteSettings(
  settings: SiteSettings,
  raw: SystemOption[] | undefined
): SiteSettings {
  if (!raw || raw.length === 0) return settings

  const optionMap = new Map(raw.map((item) => [item.key, item.value]))
  const next = { ...settings }
  const legacyMap = [
    { current: 'console_setting.announcements', legacy: 'Announcements' },
    { current: 'console_setting.api_info', legacy: 'ApiInfo' },
    { current: 'console_setting.faq', legacy: 'FAQ' },
  ] as const

  for (const { current, legacy } of legacyMap) {
    if (!optionMap.has(current)) {
      const legacyValue = optionMap.get(legacy)
      if (legacyValue !== undefined) {
        next[current] = legacyValue
      }
    }
  }

  return next
}

export function SiteSettings() {
  return (
    <SettingsPage
      routePath='/_authenticated/system-settings/site/$section'
      defaultSettings={defaultSiteSettings}
      defaultSection={SITE_DEFAULT_SECTION}
      getSectionContent={getSiteSectionContent}
      getSectionMeta={getSiteSectionMeta}
      resolveSettings={resolveSiteSettings}
    />
  )
}
