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
  Info,
  Languages,
  LayoutDashboard,
  LogOut,
  Moon,
  Sun,
  User,
} from 'lucide-react'
import { useCallback, useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { SignOutDialog } from '@/components/sign-out-dialog'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
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
import { useTheme } from '@/context/theme-provider'
import useDialogState from '@/hooks/use-dialog'
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

export function ProfileDropdown() {
  const { i18n, t } = useTranslation()
  const navigate = useNavigate()
  const [open, setOpen] = useDialogState()
  const user = useAuthStore((state) => state.auth.user)
  const { status } = useStatus()
  const navModules = parseHeaderNavModulesFromStatus(
    status as Record<string, unknown> | null
  )
  const showDocs = navModules.docs !== false
  const showAbout = navModules.about !== false
  const currentLanguage = normalizeInterfaceLanguage(i18n.language)
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
  const { displayName, roleLabel } = useUserDisplay(user)
  const { theme, setTheme } = useTheme()
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
            <Button variant='ghost' className='relative size-8 rounded-full p-0' />
          }
        >
          <Avatar className='size-8'>
            <AvatarFallback
              className={`${avatarFallbackClassName} text-[11px]`}
              style={avatarFallbackStyle}
            >
              {avatarFallback}
            </AvatarFallback>
          </Avatar>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end' sideOffset={8} className='w-56'>
          <div className='flex items-center gap-2 px-1.5 py-1.5'>
            <Avatar className='size-8'>
              <AvatarFallback
                className={`${avatarFallbackClassName} text-xs`}
                style={avatarFallbackStyle}
              >
                {avatarFallback}
              </AvatarFallback>
            </Avatar>
            <div className='flex flex-1 flex-col gap-0.5 overflow-hidden'>
              <p className='text-foreground truncate text-sm font-medium'>
                {displayName}
              </p>
              <div className='flex items-center gap-1.5'>
                <span className='text-muted-foreground text-xs'>
                  {roleLabel}
                </span>
                {user?.group && (
                  <>
                    <span className='text-muted-foreground text-xs'>·</span>
                    <span className='text-muted-foreground truncate text-xs'>
                      {String(user.group)}
                    </span>
                  </>
                )}
              </div>
            </div>
          </div>

          <DropdownMenuSeparator />

          <DropdownMenuItem onClick={() => navigate({ to: '/dashboard' })}>
            <LayoutDashboard className='size-4' />
            {t('Go to Dashboard')}
          </DropdownMenuItem>

          <DropdownMenuItem onClick={() => navigate({ to: '/profile' })}>
            <User className='size-4' />
            {t('Profile')}
          </DropdownMenuItem>

          <DropdownMenuSeparator />

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

          <DropdownMenuItem variant='destructive' onClick={() => setOpen(true)}>
            <LogOut className='size-4' />
            {t('Sign out')}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <SignOutDialog open={!!open} onOpenChange={setOpen} />
    </>
  )
}
