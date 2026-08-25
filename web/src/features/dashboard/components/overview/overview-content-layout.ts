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
import { cn } from '@/lib/utils'

export interface OverviewContentLayoutInput {
  isAdmin: boolean
  showApiInfo: boolean
  showAnnouncements: boolean
  showFaq: boolean
  showGroupAvailability: boolean
}

export interface OverviewContentLayout {
  show: boolean
  showLeft: boolean
  showPerformance: boolean
  showApiInfo: boolean
  showAnnouncements: boolean
  showFaq: boolean
  showGroupAvailability: boolean
  pairApiAndAnnouncements: boolean
  groupAvailabilityFill: boolean
  outerClassName: string
  leftClassName: string
  apiAnnouncementsRowClassName: string
  groupAvailabilityItemClassName: string
}

export function getOverviewContentLayout(
  input: OverviewContentLayoutInput
): OverviewContentLayout {
  const showLeft =
    input.isAdmin ||
    input.showApiInfo ||
    input.showAnnouncements ||
    input.showFaq
  const pairApiAndAnnouncements = input.showApiInfo && input.showAnnouncements
  const groupAvailabilityFill = showLeft && input.showGroupAvailability

  return {
    show: showLeft || input.showGroupAvailability,
    showLeft,
    showPerformance: input.isAdmin,
    showApiInfo: input.showApiInfo,
    showAnnouncements: input.showAnnouncements,
    showFaq: input.showFaq,
    showGroupAvailability: input.showGroupAvailability,
    pairApiAndAnnouncements,
    groupAvailabilityFill,
    outerClassName: cn(
      'grid grid-cols-1 gap-4',
      showLeft &&
        input.showGroupAvailability &&
        'xl:grid-cols-[minmax(0,1fr)_minmax(18rem,22rem)] xl:items-stretch'
    ),
    leftClassName: 'flex w-full min-w-0 flex-col gap-4 xl:self-start',
    apiAnnouncementsRowClassName:
      'grid min-w-0 grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1.25fr)_minmax(0,1fr)]',
    groupAvailabilityItemClassName: groupAvailabilityFill
      ? 'flex min-h-80 min-w-0 flex-col xl:h-full'
      : 'min-w-0',
  }
}
