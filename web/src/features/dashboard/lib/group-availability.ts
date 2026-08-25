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
export function parseGroupAvailabilityGroups(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.filter(
      (item): item is string => typeof item === 'string' && item.length > 0
    )
  }
  if (typeof value !== 'string' || value.trim() === '') return []
  try {
    const parsed: unknown = JSON.parse(value)
    if (!Array.isArray(parsed)) return []
    return parsed.filter(
      (item): item is string => typeof item === 'string' && item.length > 0
    )
  } catch {
    return []
  }
}

export function filterAvailabilityGroups<T extends { group: string }>(
  groups: T[],
  selected: string[]
): T[] {
  if (selected.length === 0) return []
  const selectedSet = new Set(selected)
  return groups.filter((item) => selectedSet.has(item.group))
}

export function bucketSecondsToMinutes(
  seconds: number | null | undefined
): number {
  if (seconds == null || !Number.isFinite(seconds) || seconds < 1) return 60
  return Math.max(1, Math.round(seconds / 60))
}

export function isHourAvailabilityWindow(minutes: number): boolean {
  return minutes === 60
}
