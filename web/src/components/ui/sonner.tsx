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
'use client'

import {
  Alert02Icon,
  CheckmarkCircle02Icon,
  InformationCircleIcon,
  Loading03Icon,
  MultiplicationSignCircleIcon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useRouterState } from '@tanstack/react-router'
import { Toaster as Sonner, type ToasterProps } from 'sonner'

import { useTheme } from '@/context/theme-provider'

const AUTH_HEADER_OFFSET = '4.25rem'
const CONSOLE_HEADER_OFFSET = '4.75rem'

const Toaster = (props: ToasterProps) => {
  const { resolvedTheme } = useTheme()
  const isAuthPage = useRouterState({
    select: (state) => {
      const pathname = state.location.pathname
      return (
        pathname === '/sign-in' ||
        pathname === '/sign-up' ||
        pathname === '/register' ||
        pathname === '/forgot-password' ||
        pathname === '/reset' ||
        pathname === '/otp' ||
        pathname.startsWith('/oauth') ||
        pathname.startsWith('/user/reset')
      )
    },
  })
  const headerOffset = isAuthPage ? AUTH_HEADER_OFFSET : CONSOLE_HEADER_OFFSET

  return (
    <Sonner
      theme={resolvedTheme}
      className='toaster group'
      position='top-center'
      closeButton={false}
      duration={5000}
      gap={8}
      offset={{ top: headerOffset }}
      mobileOffset={{ top: headerOffset }}
      icons={{
        success: (
          <HugeiconsIcon
            icon={CheckmarkCircle02Icon}
            strokeWidth={2}
            className='size-3.5'
          />
        ),
        info: (
          <HugeiconsIcon
            icon={InformationCircleIcon}
            strokeWidth={2}
            className='size-3.5'
          />
        ),
        warning: (
          <HugeiconsIcon
            icon={Alert02Icon}
            strokeWidth={2}
            className='size-3.5'
          />
        ),
        error: (
          <HugeiconsIcon
            icon={MultiplicationSignCircleIcon}
            strokeWidth={2}
            className='size-3.5'
          />
        ),
        loading: (
          <HugeiconsIcon
            icon={Loading03Icon}
            strokeWidth={2}
            className='size-3.5 animate-spin'
          />
        ),
      }}
      style={
        {
          '--normal-bg': 'var(--popover)',
          '--normal-text': 'var(--popover-foreground)',
          '--normal-border': 'var(--border)',
          '--border-radius': 'var(--radius)',
          '--width': '100%',
        } as React.CSSProperties
      }
      {...props}
    />
  )
}

export { Toaster }
