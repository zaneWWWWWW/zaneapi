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
import { getLobeIcon } from '@/lib/lobe-icon'
import { cn } from '@/lib/utils'

interface IconCardProps {
  iconName: string
  size?: number
  className?: string
}

/**
 * Reusable icon card for the home gateway diagram.
 */
export function IconCard({ iconName, size = 32, className }: IconCardProps) {
  return (
    <div
      className={cn(
        'border-border bg-card relative overflow-hidden rounded-lg border p-5 shadow-xs',
        className
      )}
    >
      <div className='flex items-center justify-center'>
        {getLobeIcon(iconName, size)}
      </div>
    </div>
  )
}
