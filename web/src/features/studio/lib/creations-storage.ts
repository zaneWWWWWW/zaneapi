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

import type { CreationItem } from '../types'

export const INTERRUPTED_GENERATION_MESSAGE =
  'Generation was interrupted. Please try again.'

function isPendingStatus(status: CreationItem['status']) {
  return status === 'queued' || status === 'in_progress'
}

function hasRenderableImage(item: CreationItem) {
  return Boolean(item.url || item.b64Json)
}

export function reviveSavedCreations(items: CreationItem[]): CreationItem[] {
  return items.map((item) => {
    if (item.type === 'video' && item.taskId && isPendingStatus(item.status)) {
      return item
    }

    if (isPendingStatus(item.status)) {
      return {
        ...item,
        status: 'failed',
        failReason: INTERRUPTED_GENERATION_MESSAGE,
      }
    }

    if (item.type === 'image' && item.status === 'completed' && !hasRenderableImage(item)) {
      return {
        ...item,
        status: 'failed',
        failReason: INTERRUPTED_GENERATION_MESSAGE,
      }
    }

    return item
  })
}
