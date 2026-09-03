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
import type { ImgHTMLAttributes } from 'react'

import { DEFAULT_LOGO } from '@/lib/constants'
import { cn } from '@/lib/utils'

/**
 * Renders a system logo while keeping the bundled mark legible in dark mode.
 * Custom administrator-provided logo URLs are left unchanged.
 */
export function BrandImage(props: ImgHTMLAttributes<HTMLImageElement>) {
  const { className, ...imageProps } = props
  return (
    <img
      {...imageProps}
      className={cn(className, props.src === DEFAULT_LOGO && 'dark:invert')}
    />
  )
}
