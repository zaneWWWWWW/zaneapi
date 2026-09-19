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

import { useEffect, useRef } from 'react'

import { fetchVideoTask } from '../api'
import type { CreationItem } from '../types'

interface UseTaskPollerOptions {
  creations: CreationItem[]
  updateCreation: (id: string, patch: Partial<CreationItem>) => void
}

export function useTaskPoller(options: UseTaskPollerOptions) {
  const { creations, updateCreation } = options
  const updateCreationRef = useRef(updateCreation)
  updateCreationRef.current = updateCreation

  const pendingVideos = creations.filter(
    (c) =>
      c.type === 'video' &&
      (c.status === 'queued' || c.status === 'in_progress') &&
      (c.taskId || c.id)
  )

  const pendingIdsKey = pendingVideos.map((c) => c.taskId || c.id).join(',')

  useEffect(() => {
    if (pendingVideos.length === 0) return

    let isMounted = true
    const pollInterval = setInterval(async () => {
      for (const item of pendingVideos) {
        const taskId = item.taskId
        if (!taskId) continue
        try {
          const res = await fetchVideoTask(taskId)
          if (!isMounted) return

          const anyRes = res as unknown as {
            status?: string
            url?: string
            data?: Array<{ url?: string }>
            error?: { message?: string }
            progress?: number
          }
          const statusStr = (anyRes.status || '').toLowerCase()
          if (
            statusStr === 'completed' ||
            statusStr === 'success' ||
            statusStr === 'succeeded' ||
            (Array.isArray(anyRes.data) && anyRes.data[0]?.url)
          ) {
            updateCreationRef.current(item.id, {
              status: 'completed',
              progress: 100,
              url:
                anyRes.url ||
                anyRes.data?.[0]?.url ||
                `/v1/videos/${taskId}/content`,
            })
          } else if (statusStr === 'failed' || statusStr === 'error') {
            updateCreationRef.current(item.id, {
              status: 'failed',
              failReason:
                anyRes.error?.message || 'Video generation failed or timed out',
            })
          } else {
            // Still in progress / queued
            let progress = 5
            if (typeof anyRes.progress === 'number') {
              progress = anyRes.progress
            } else if (statusStr === 'in_progress') {
              progress = Math.max(item.progress || 10, 20)
            }

            updateCreationRef.current(item.id, {
              status: statusStr === 'in_progress' ? 'in_progress' : 'queued',
              progress,
            })
          }
        } catch {
          // Keep polling, ignore transient fetch errors
        }
      }
    }, 3000)

    return () => {
      isMounted = false
      clearInterval(pollInterval)
    }
  }, [pendingIdsKey, pendingVideos])
}
