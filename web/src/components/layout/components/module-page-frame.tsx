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
import { PageTransition } from '@/components/page-transition'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

type ModulePageFrameProps = {
  children: React.ReactNode
  className?: string
  contentClassName?: string
}

export function ModulePageFrame(props: ModulePageFrameProps) {
  const user = useAuthStore((state) => state.auth.user)

  return (
    <div
      className={cn(
        'relative',
        user && 'min-h-0 flex-1 overflow-y-auto',
        props.className
      )}
    >
      <PageTransition
        className={cn(
          'relative mx-auto w-full px-3 pb-8 sm:px-6 sm:pb-10 xl:px-8',
          user ? 'pt-4 sm:pt-6' : 'pt-16 sm:pt-20',
          props.contentClassName
        )}
      >
        {props.children}
      </PageTransition>
    </div>
  )
}
