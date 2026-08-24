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
import { Link } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { LanguageSwitcher } from '@/components/language-switcher'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

import { APP_HEADER_CHROME } from '../constants'
import type { TopNavLink } from '../types'
import { HeaderLogo } from './header-logo'

export interface PublicHeaderProps {
  navLinks?: TopNavLink[]
  mobileLinks?: TopNavLink[]
  navContent?: React.ReactNode
  showThemeSwitch?: boolean
  showLanguageSwitcher?: boolean
  logo?: React.ReactNode
  siteName?: string
  homeUrl?: string
  leftContent?: React.ReactNode
  rightContent?: React.ReactNode
  showNavigation?: boolean
  showAuthButtons?: boolean
  showNotifications?: boolean
  className?: string
}

export function PublicHeader(props: PublicHeaderProps) {
  const {
    showThemeSwitch = true,
    showLanguageSwitcher = true,
    logo: customLogo,
    siteName: customSiteName,
    homeUrl = '/',
    showAuthButtons = true,
  } = props

  const { t } = useTranslation()
  const [scrolled, setScrolled] = useState(false)
  const { auth } = useAuthStore()
  const {
    systemName,
    logo: systemLogo,
    loading,
    logoLoaded,
  } = useSystemConfig()

  const isAuthenticated = !!auth.user
  const displaySiteName = customSiteName || systemName

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  let brandLogo: React.ReactNode
  if (loading) {
    brandLogo = <Skeleton className='size-full rounded-lg' />
  } else if (customLogo) {
    brandLogo = customLogo
  } else {
    brandLogo = (
      <HeaderLogo
        src={systemLogo}
        loading={loading}
        logoLoaded={logoLoaded}
        className='size-full rounded-lg object-contain'
      />
    )
  }

  let accountControl: React.ReactNode
  if (loading) {
    accountControl = <Skeleton className='size-8 rounded-full' />
  } else if (isAuthenticated) {
    accountControl = <ProfileDropdown />
  } else {
    accountControl = (
      <Button
        size='sm'
        className='h-8 rounded-full px-4 text-xs font-medium'
        render={<Link to='/sign-in' />}
      >
        {t('Sign in')}
      </Button>
    )
  }

  return (
    <header className='pointer-events-none fixed inset-x-0 top-0 z-50'>
      <div
        className={cn(
          'pointer-events-auto border-b transition-colors duration-200',
          scrolled
            ? 'bg-background/90 border-border backdrop-blur-xl'
            : 'bg-background/75 border-transparent backdrop-blur-md'
        )}
      >
        <div className='mx-auto flex h-14 max-w-7xl items-center justify-between px-4 md:px-6'>
          {props.leftContent ?? (
            <Link to={homeUrl} className='flex shrink-0 items-center gap-2.5'>
              <div className='flex size-7 shrink-0 items-center justify-center'>
                {brandLogo}
              </div>
              <span className='text-sm font-semibold tracking-tight'>
                {loading ? <Skeleton className='h-4 w-16' /> : displaySiteName}
              </span>
            </Link>
          )}

          {props.rightContent ?? (
            <div className='flex items-center gap-1'>
              {APP_HEADER_CHROME.publicNavLinks ? props.navContent : null}
              {showLanguageSwitcher && !isAuthenticated && <LanguageSwitcher />}
              {showThemeSwitch && !isAuthenticated && <ThemeSwitch />}
              {showAuthButtons && accountControl}
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
