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
import {
  BookOpen,
  Check,
  ChevronRight,
  Info,
  Languages,
  LogOut,
  Moon,
  Sun,
  User,
} from 'lucide-react'
import { useCallback, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { SignOutDialog } from '@/components/sign-out-dialog'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useSidebar } from '@/components/ui/sidebar'
import { useTheme } from '@/context/theme-provider'
import { useStatus } from '@/hooks/use-status'
import { useUserDisplay } from '@/hooks/use-user-display'
import {
  INTERFACE_LANGUAGE_OPTIONS,
  normalizeInterfaceLanguage,
} from '@/i18n/languages'
import { api } from '@/lib/api'
import { getUserAvatarFallback, getUserAvatarStyle } from '@/lib/avatar'
import { parseHeaderNavModulesFromStatus } from '@/lib/nav-modules'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

const avatarFallbackClassName = 'font-semibold text-white'

export function SidebarAccount() {
  const { i18n, t } = useTranslation()
  const navigate = useNavigate()
  const { isMobile, state } = useSidebar()
  const [signOutOpen, setSignOutOpen] = useState(false)
  const { theme, setTheme } = useTheme()
  const user = useAuthStore((state) => state.auth.user)
  const { displayName } = useUserDisplay(user)
  const { status } = useStatus()
  const currentLanguage = normalizeInterfaceLanguage(i18n.language)
  const collapsed = !isMobile && state === 'collapsed'
  const navModules = parseHeaderNavModulesFromStatus(
    status as Record<string, unknown> | null
  )
  const showDocs = navModules.docs !== false
  const showAbout = navModules.about !== false

  const handleChangeLanguage = useCallback(
    async (code: string) => {
      await i18n.changeLanguage(code)
      if (user) {
        try {
          await api.put('/api/user/self', { language: code })
        } catch {
          // Best-effort persistence; don't block the UI on failure
        }
      }
    },
    [i18n, user]
  )

  const avatarName = user?.username || displayName
  const avatarFallback = getUserAvatarFallback(avatarName)
  const avatarFallbackStyle = useMemo(
    () => getUserAvatarStyle(avatarName),
    [avatarName]
  )

  return (
    <>
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger
          render={
            <button
              type='button'
              className={cn(
                'hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex w-full items-center gap-2 rounded-lg p-2 text-left outline-none',
                'focus-visible:ring-sidebar-ring focus-visible:ring-2',
                collapsed && 'size-9 justify-center p-0'
              )}
              aria-label={t('Personal account')}
            />
          }
        >
          <Avatar className='size-7 shrink-0'>
            <AvatarFallback
              className={`${avatarFallbackClassName} text-[11px]`}
              style={avatarFallbackStyle}
            >
              {avatarFallback}
            </AvatarFallback>
          </Avatar>
          <span className={cn('min-w-0 flex-1', collapsed && 'hidden')}>
            <span className='text-sidebar-foreground block truncate text-sm font-medium'>
              {displayName}
            </span>
            <span className='text-muted-foreground block truncate text-xs'>
              {t('Personal account')}
            </span>
          </span>
          <ChevronRight
            className={cn(
              'text-muted-foreground size-4 shrink-0',
              collapsed && 'hidden'
            )}
          />
        </DropdownMenuTrigger>
        <DropdownMenuContent
          side='top'
          align='start'
          sideOffset={8}
          className='w-60'
        >
          <div className='flex items-center gap-2 px-1.5 py-1.5'>
            <Avatar className='size-8'>
              <AvatarFallback
                className={`${avatarFallbackClassName} text-xs`}
                style={avatarFallbackStyle}
              >
                {avatarFallback}
              </AvatarFallback>
            </Avatar>
            <div className='min-w-0 flex-1'>
              <p className='text-foreground truncate text-sm font-medium'>
                {displayName}
              </p>
              <p className='text-muted-foreground truncate text-xs'>
                {t('Personal account')}
              </p>
            </div>
          </div>

          <DropdownMenuSeparator />

          <DropdownMenuItem onClick={() => navigate({ to: '/profile' })}>
            <User className='size-4' />
            {t('Profile')}
          </DropdownMenuItem>

          <DropdownMenuSub>
            <DropdownMenuSubTrigger>
              {theme === 'dark' ? (
                <Moon className='size-4' />
              ) : (
                <Sun className='size-4' />
              )}
              {t('Change theme')}
            </DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
              <DropdownMenuItem onClick={() => setTheme('light')}>
                <Sun className='size-4' />
                {t('Light')}
                <Check
                  size={14}
                  className={cn('ms-auto', theme !== 'light' && 'hidden')}
                />
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTheme('dark')}>
                <Moon className='size-4' />
                {t('Dark')}
                <Check
                  size={14}
                  className={cn('ms-auto', theme !== 'dark' && 'hidden')}
                />
              </DropdownMenuItem>
            </DropdownMenuSubContent>
          </DropdownMenuSub>

          <DropdownMenuSub>
            <DropdownMenuSubTrigger>
              <Languages className='size-4' />
              {t('Change language')}
            </DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
              {INTERFACE_LANGUAGE_OPTIONS.map((lang) => (
                <DropdownMenuItem
                  key={lang.code}
                  onClick={() => handleChangeLanguage(lang.code)}
                >
                  {lang.label}
                  <Check
                    size={14}
                    className={cn(
                      'ms-auto',
                      currentLanguage !== lang.code && 'hidden'
                    )}
                  />
                </DropdownMenuItem>
              ))}
            </DropdownMenuSubContent>
          </DropdownMenuSub>

          {(showDocs || showAbout) && (
            <>
              <DropdownMenuSeparator />
              {showDocs && (
                <DropdownMenuItem onClick={() => navigate({ to: '/docs' })}>
                  <BookOpen className='size-4' />
                  {t('Docs')}
                </DropdownMenuItem>
              )}
              {showAbout && (
                <DropdownMenuItem onClick={() => navigate({ to: '/about' })}>
                  <Info className='size-4' />
                  {t('About')}
                </DropdownMenuItem>
              )}
            </>
          )}

          <DropdownMenuSeparator />

          <DropdownMenuItem
            variant='destructive'
            onClick={() => setSignOutOpen(true)}
          >
            <LogOut className='size-4' />
            {t('Sign out')}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <SignOutDialog open={!!signOutOpen} onOpenChange={setSignOutOpen} />
    </>
  )
}
