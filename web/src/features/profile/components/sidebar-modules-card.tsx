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
import { LayoutDashboard } from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { IconBadge } from '@/components/ui/icon-badge'
import { Switch } from '@/components/ui/switch'
import {
  parseSidebarModulesAdmin,
  type SidebarModulesAdminConfig,
} from '@/features/system-settings/maintenance/config'
import { useStatus } from '@/hooks/use-status'
import { api } from '@/lib/api'
import { parseHeaderNavModulesFromStatus } from '@/lib/nav-modules'
import { useAuthStore } from '@/stores/auth-store'

type SidebarModuleConfig = {
  enabled: boolean
  [key: string]: boolean
}

type SidebarModulesConfig = Record<string, SidebarModuleConfig>

type SidebarItemDef = {
  section: string
  key: string
  title: string
  description: string
  extraKeys?: string[]
}

function ensureSection(
  config: SidebarModulesConfig,
  section: string
): SidebarModuleConfig {
  const current = config[section]
  if (!current) return { enabled: true }
  return { ...current, enabled: current.enabled !== false }
}

function isSiteModuleEnabled(
  admin: SidebarModulesAdminConfig,
  section: string,
  keys: string[]
): boolean {
  const siteSection = admin[section]
  if (!siteSection || siteSection.enabled === false) return false
  return keys.some((key) => siteSection[key] === true)
}

export function SidebarModulesCard() {
  const { t } = useTranslation()
  const { status } = useStatus()
  const [loading, setLoading] = useState(false)
  const [config, setConfig] = useState<SidebarModulesConfig>({})
  const currentUser = useAuthStore((s) => s.auth.user)
  const setUser = useAuthStore((s) => s.auth.setUser)
  const headerNav = parseHeaderNavModulesFromStatus(
    status as Record<string, unknown> | null
  )
  const siteModules = useMemo(
    () =>
      parseSidebarModulesAdmin(
        status?.SidebarModulesAdmin as string | null | undefined
      ),
    [status?.SidebarModulesAdmin]
  )

  const items = useMemo<SidebarItemDef[]>(() => {
    const catalog: (SidebarItemDef & { siteEnabled: boolean })[] = [
      {
        section: 'console',
        key: 'pricing',
        title: t('Model Square'),
        description: t('Browse models and prices'),
        siteEnabled:
          headerNav.pricing.enabled &&
          isSiteModuleEnabled(siteModules, 'console', ['pricing']),
      },
      {
        section: 'console',
        key: 'rankings',
        title: t('Rankings'),
        description: t('View model rankings'),
        siteEnabled:
          headerNav.rankings.enabled &&
          isSiteModuleEnabled(siteModules, 'console', ['rankings']),
      },
      {
        section: 'chat',
        key: 'playground',
        title: t('Playground'),
        description: t('AI model testing environment'),
        siteEnabled: isSiteModuleEnabled(siteModules, 'chat', ['playground']),
      },
      {
        section: 'chat',
        key: 'chat',
        title: t('Chat'),
        description: t('Chat session management'),
        siteEnabled: isSiteModuleEnabled(siteModules, 'chat', ['chat']),
      },
      {
        section: 'console',
        key: 'detail',
        title: t('Dashboard'),
        description: t('System data statistics'),
        siteEnabled: isSiteModuleEnabled(siteModules, 'console', ['detail']),
      },
      {
        section: 'console',
        key: 'token',
        title: t('API Keys'),
        description: t('API token management'),
        siteEnabled: isSiteModuleEnabled(siteModules, 'console', ['token']),
      },
      {
        section: 'console',
        key: 'log',
        title: t('Usage Logs'),
        description: t('API usage records'),
        extraKeys: ['midjourney', 'task'],
        siteEnabled: isSiteModuleEnabled(siteModules, 'console', [
          'log',
          'midjourney',
          'task',
        ]),
      },
      {
        section: 'personal',
        key: 'topup',
        title: t('Wallet'),
        description: t('Balance and top-up management'),
        siteEnabled: isSiteModuleEnabled(siteModules, 'personal', ['topup']),
      },
    ]

    return catalog
      .filter((item) => item.siteEnabled)
      .map(({ siteEnabled: _siteEnabled, ...item }) => item)
  }, [headerNav.pricing.enabled, headerNav.rankings.enabled, siteModules, t])

  const applyItemDefaults = useCallback(
    (base: SidebarModulesConfig): SidebarModulesConfig => {
      const next: SidebarModulesConfig = { ...base }
      for (const item of items) {
        const section = ensureSection(next, item.section)
        section[item.key] = true
        for (const extraKey of item.extraKeys ?? []) {
          section[extraKey] = true
        }
        next[item.section] = section
      }
      const personal = ensureSection(next, 'personal')
      personal.personal = true
      next.personal = personal
      return next
    },
    [items]
  )

  const loadConfig = useCallback(async () => {
    try {
      const res = await api.get('/api/user/self')
      if (res.data.success && res.data.data?.sidebar_modules) {
        const raw = res.data.data.sidebar_modules
        const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
        setConfig(parsed && typeof parsed === 'object' ? parsed : {})
        return
      }
      setConfig(applyItemDefaults({}))
    } catch {
      /* ignore */
    }
  }, [applyItemDefaults])

  useEffect(() => {
    void loadConfig()
  }, [loadConfig])

  const toggleItem = (item: SidebarItemDef, val: boolean) => {
    setConfig((prev) => {
      const section = ensureSection(prev, item.section)
      section[item.key] = val
      for (const extraKey of item.extraKeys ?? []) {
        section[extraKey] = val
      }
      return { ...prev, [item.section]: section }
    })
  }

  const handleSave = async () => {
    setLoading(true)
    try {
      const next = { ...config }
      for (const item of items) {
        const section = ensureSection(next, item.section)
        const value = section[item.key] !== false
        section[item.key] = value
        for (const extraKey of item.extraKeys ?? []) {
          section[extraKey] = value
        }
        next[item.section] = section
      }
      const personal = ensureSection(next, 'personal')
      personal.personal = true
      next.personal = personal

      const serialized = JSON.stringify(next)
      const res = await api.put('/api/user/self', {
        sidebar_modules: serialized,
      })
      if (res.data.success) {
        setConfig(next)
        if (currentUser) {
          setUser({ ...currentUser, sidebar_modules: serialized })
        }
        toast.success(t('Saved successfully'))
      } else {
        toast.error(res.data.message || t('Save failed'))
      }
    } catch {
      toast.error(t('Save failed, please retry'))
    } finally {
      setLoading(false)
    }
  }

  const handleReset = () => {
    setConfig((prev) => applyItemDefaults(prev))
    toast.success(t('Reset to default configuration'))
  }

  return (
    <Card data-card-hover='false' className='gap-0 overflow-hidden py-0'>
      <CardHeader className='border-b p-3 !pb-3 sm:p-5 sm:!pb-5'>
        <div className='flex items-center gap-3'>
          <IconBadge tone='info' size='title'>
            <LayoutDashboard />
          </IconBadge>
          <div className='min-w-0'>
            <CardTitle className='text-lg tracking-tight sm:text-xl'>
              {t('Sidebar Personal Settings')}
            </CardTitle>
            <CardDescription className='text-xs sm:text-sm'>
              {t('Customize sidebar display content')}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className='space-y-4 p-3 sm:space-y-5 sm:p-5'>
        <div className='grid grid-cols-1 gap-2'>
          {items.map((item) => (
            <div
              key={`${item.section}.${item.key}`}
              className='flex min-h-16 items-center justify-between rounded-lg border p-3'
            >
              <div className='mr-2 min-w-0'>
                <p className='truncate text-sm font-medium'>{item.title}</p>
                <p className='text-muted-foreground truncate text-xs'>
                  {item.description}
                </p>
              </div>
              <Switch
                checked={config[item.section]?.[item.key] !== false}
                onCheckedChange={(v) => toggleItem(item, v)}
              />
            </div>
          ))}
        </div>

        <div className='flex flex-col-reverse gap-2 border-t pt-4 sm:flex-row sm:justify-end'>
          <Button variant='outline' onClick={handleReset}>
            {t('Reset to Default')}
          </Button>
          <Button onClick={handleSave} disabled={loading}>
            {loading ? t('Saving...') : t('Save Changes')}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
