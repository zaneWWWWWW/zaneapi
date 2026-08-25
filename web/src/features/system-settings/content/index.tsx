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
import { useNavigate } from '@tanstack/react-router'

import { SettingsPage } from '../components/settings-page'
import type { ContentSettings, SystemOption } from '../types'
import {
  CONTENT_DEFAULT_SECTION,
  CONTENT_SECTION_IDS,
  getContentSectionContent,
  getContentSectionMeta,
  type ContentSectionId,
} from './section-registry.tsx'

const defaultContentSettings: ContentSettings = {
  'console_setting.uptime_kuma_groups': '[]',
  'console_setting.uptime_kuma_enabled': false,
  'console_setting.group_availability_groups': '[]',
  DataExportEnabled: false,
  DataExportDefaultTime: 'hour',
  DataExportInterval: 5,
  Chats: '[]',
  DrawingEnabled: false,
  MjNotifyEnabled: false,
  MjAccountFilterEnabled: false,
  MjForwardUrlEnabled: false,
  MjModeClearEnabled: false,
  MjActionCheckSuccessEnabled: false,
}

function resolveContentSettings(
  settings: ContentSettings,
  raw: SystemOption[] | undefined
): ContentSettings {
  if (!raw || raw.length === 0) return settings

  const optionMap = new Map(raw.map((item) => [item.key, item.value]))
  const next = { ...settings }

  if (!optionMap.has('console_setting.uptime_kuma_groups')) {
    const legacyUrl = optionMap.get('UptimeKumaUrl')
    const legacySlug = optionMap.get('UptimeKumaSlug')
    if (legacyUrl && legacySlug) {
      next['console_setting.uptime_kuma_groups'] = JSON.stringify([
        { id: 1, categoryName: 'Legacy', url: legacyUrl, slug: legacySlug },
      ])
    }
  }

  return next
}

export function ContentSettings() {
  const navigate = useNavigate()
  return (
    <SettingsPage
      routePath='/_authenticated/system-settings/content/$section'
      defaultSettings={defaultContentSettings}
      defaultSection={CONTENT_DEFAULT_SECTION}
      getSectionContent={getContentSectionContent}
      getSectionMeta={getContentSectionMeta}
      loadingMessage='Loading content settings...'
      resolveSettings={resolveContentSettings}
      pageTitleKey='Console display'
      sectionTabs={CONTENT_SECTION_IDS.map((id) => ({
        id,
        titleKey: getContentSectionMeta(id).titleKey,
      }))}
      onSectionChange={(section: ContentSectionId) => {
        void navigate({
          to: '/system-settings/content/$section',
          params: { section },
        })
      }}
    />
  )
}
