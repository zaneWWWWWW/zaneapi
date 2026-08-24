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
export const HIDDEN_SYNC_CHANNEL_STORAGE_KEY = 'ratio-sync-hidden-channel-ids'

export function parseHiddenSyncChannelIds(raw: string | null): number[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return [
      ...new Set(
        parsed.filter(
          (id): id is number => typeof id === 'number' && Number.isInteger(id)
        )
      ),
    ]
  } catch {
    return []
  }
}

export function serializeHiddenSyncChannelIds(ids: number[]): string {
  return JSON.stringify([...new Set(ids.filter((id) => Number.isInteger(id)))])
}

export function loadHiddenSyncChannelIds(): number[] {
  if (typeof localStorage === 'undefined') return []
  try {
    return parseHiddenSyncChannelIds(
      localStorage.getItem(HIDDEN_SYNC_CHANNEL_STORAGE_KEY)
    )
  } catch {
    return []
  }
}

export function saveHiddenSyncChannelIds(ids: number[]): void {
  if (typeof localStorage === 'undefined') return
  try {
    localStorage.setItem(
      HIDDEN_SYNC_CHANNEL_STORAGE_KEY,
      serializeHiddenSyncChannelIds(ids)
    )
  } catch {
    // Ignore quota / private-mode failures; the in-memory list still applies.
  }
}

export function hideSyncChannelId(ids: number[], channelId: number): number[] {
  if (ids.includes(channelId)) return ids
  return [...ids, channelId]
}

export function filterVisibleSyncChannels<T extends { id: number }>(
  channels: T[],
  hiddenIds: number[]
): T[] {
  if (hiddenIds.length === 0) return channels
  const hidden = new Set(hiddenIds)
  return channels.filter((channel) => !hidden.has(channel.id))
}
