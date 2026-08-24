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
interface FeatureItemProps {
  title: string
  description: string
  icon: React.ReactNode
}

/**
 * Individual feature item with icon, title, and description
 */
export function FeatureItem({ title, description, icon }: FeatureItemProps) {
  return (
    <div className='group/feature text-foreground flex flex-col gap-4 p-4'>
      {/* Icon */}
      <div className='flex items-center self-start'>
        <div className='bg-muted text-muted-foreground group-hover/feature:text-foreground flex size-10 items-center justify-center rounded-lg border shadow-xs transition-all duration-300 group-hover/feature:scale-105'>
          {icon}
        </div>
      </div>
      {/* Title */}
      <h3 className='text-sm leading-none font-semibold tracking-tight sm:text-base'>
        {title}
      </h3>
      {/* Description */}
      <p className='text-muted-foreground max-w-[240px] text-sm text-balance'>
        {description}
      </p>
    </div>
  )
}
