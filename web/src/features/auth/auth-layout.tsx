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
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'

type AuthLayoutProps = {
  children: React.ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()

  return (
    <div className='bg-background flex min-h-svh flex-col'>
      <header className='flex h-14 items-center border-b px-4 sm:px-8'>
        <Link
          to='/'
          className='flex items-center gap-2 transition-opacity hover:opacity-80'
        >
          <div className='relative size-7'>
            {loading ? (
              <Skeleton className='absolute inset-0 rounded-md' />
            ) : (
              <img
                src={logo}
                alt={t('Logo')}
                className='size-7 rounded-md object-cover'
              />
            )}
          </div>
          {loading ? (
            <Skeleton className='h-4 w-20' />
          ) : (
            <span className='text-sm font-medium'>{systemName}</span>
          )}
        </Link>
      </header>
      <div className='flex flex-1 items-center justify-center px-4 py-10'>
        <div className='w-full max-w-[400px]'>{children}</div>
      </div>
    </div>
  )
}
