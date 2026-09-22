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

import { getPerfMetricsSummary } from '@/features/performance-metrics/api'
import { getPricing } from '@/features/pricing/api'
import { api } from '@/lib/api'

import { POPULAR_IMAGE_MODELS, POPULAR_VIDEO_MODELS } from './constants'

import {
  buildImageRequestBody,
  collectStudioReferenceImages,
  referenceImageFilename,
} from './lib/image-request'
import { attachModelSuccessRates } from './lib/model-success-rate'
import { buildStudioImageOptions } from './lib/studio-models'
import type {
  GroupOption,
  ImageGenerationRequest,
  ModelOption,
  OpenAIImageResponse,
  OpenAIVideoResponse,
  VideoGenerationRequest,
} from './types'

export async function generateImage(
  payload: ImageGenerationRequest,
  signal?: AbortSignal
): Promise<OpenAIImageResponse> {
  const headers: Record<string, string> = {}
  if (payload.group) {
    headers['New-Api-Group'] = payload.group
  }

  const imageFiles =
    payload.imageFiles?.filter((file) => file && file.size > 0) ?? []
  const referenceImages =
    imageFiles.length > 0 ? [] : collectStudioReferenceImages(payload)
  const referenceCount = imageFiles.length > 0 ? imageFiles.length : referenceImages.length

  // If reference images exist, convert data URLs to blobs and submit to /pg/images/edits
  if (referenceCount > 0) {
    const formData = new FormData()
    formData.append('prompt', payload.prompt)
    formData.append('model', payload.model)
    if (payload.group) formData.append('group', payload.group)
    if (payload.size) formData.append('size', payload.size)
    if (payload.n) formData.append('n', String(payload.n))

    const fieldName = referenceCount > 1 ? 'image[]' : 'image'
    headers['X-Studio-Reference-Count'] = String(referenceCount)
    if (imageFiles.length > 0) {
      for (const [index, file] of imageFiles.entries()) {
        formData.append(
          fieldName,
          file,
          file.name || referenceImageFilename(index, file.type)
        )
      }
    } else {
      for (const [index, src] of referenceImages.entries()) {
        const response = await fetch(src)
        const blob = await response.blob()
        formData.append(fieldName, blob, referenceImageFilename(index, blob.type))
      }
    }

    const res = await api.post('/pg/images/edits', formData, {
      signal,
      headers,
      skipErrorHandler: true,
    } as Record<string, unknown>)
    return res.data
  }

  const res = await api.post(
    '/pg/images/generations',
    buildImageRequestBody(payload),
    {
      signal,
      headers,
      skipErrorHandler: true,
    } as Record<string, unknown>
  )
  return res.data
}

export async function generateVideo(
  payload: VideoGenerationRequest,
  signal?: AbortSignal
): Promise<OpenAIVideoResponse> {
  const headers: Record<string, string> = {}
  if (payload.group) {
    headers['New-Api-Group'] = payload.group
  }

  const res = await api.post(
    '/pg/videos',
    {
      model: payload.model,
      prompt: payload.prompt,
      duration: payload.duration || 5,
      aspect_ratio: payload.aspect_ratio || '16:9',
      prompt_image: payload.prompt_image,
      image: payload.image,
      last_frame_image: payload.last_frame_image,
      camera_motion: payload.camera_motion,
      resolution: payload.resolution,
    },
    {
      signal,
      headers,
      skipErrorHandler: true,
    } as Record<string, unknown>
  )
  return res.data
}

export async function fetchVideoTask(
  taskId: string,
  signal?: AbortSignal
): Promise<OpenAIVideoResponse> {
  const res = await api.get(`/pg/videos/${taskId}`, {
    signal,
    skipErrorHandler: true,
  } as Record<string, unknown>)
  return res.data
}

export async function getUserAvailableModels(
  group: string
): Promise<ModelOption[]> {
  const res = await api.get('/api/user/models', {
    params: { group },
  })
  const { data } = res

  if (!data.success || !Array.isArray(data.data)) {
    return []
  }

  const rawModels: string[] = data.data
  return rawModels.map((model) => {
    const isImage =
      (POPULAR_IMAGE_MODELS as readonly string[]).includes(model) ||
      model.includes('image') ||
      model.includes('flux') ||
      model.includes('dall-e') ||
      model.includes('recraft') ||
      model.includes('ideogram') ||
      model.includes('sdxl') ||
      model.includes('stable-diffusion')

    const isVideo =
      (POPULAR_VIDEO_MODELS as readonly string[]).includes(model) ||
      model.includes('video') ||
      model.includes('kling') ||
      model.includes('sora') ||
      model.includes('grok-imagine') ||
      model.includes('hailuo') ||
      model.includes('minimax') ||
      model.includes('h3')

    let type: 'image' | 'video' | 'all' = 'all'
    if (isImage && !isVideo) type = 'image'
    else if (isVideo && !isImage) type = 'video'

    return {
      label: model,
      value: model,
      type,
      isPopular:
        (POPULAR_IMAGE_MODELS as readonly string[]).includes(model) ||
        (POPULAR_VIDEO_MODELS as readonly string[]).includes(model),
    }
  })
}

export async function loadStudioImageModels(): Promise<{
  groups: GroupOption[]
  models: ModelOption[]
}> {
  const groups = await getUserGroups()
  const [pricing, summary, modelsByGroup] = await Promise.all([
    getPricing().catch(() => null),
    getPerfMetricsSummary(24).catch(() => null),
    Promise.all(
      groups.map(async (group) => ({
        group: group.value,
        models: await getUserAvailableModels(group.value),
      }))
    ),
  ])

  const options = attachModelSuccessRates(
    buildStudioImageOptions({
      groups,
      modelsByGroup,
      pricingModels: pricing?.data ?? [],
      groupRatio: {
        ...Object.fromEntries(groups.map((group) => [group.value, group.ratio])),
        ...pricing?.group_ratio,
      },
    }),
    summary?.data.models ?? []
  )

  return { groups, models: options }
}

export async function getUserGroups(): Promise<GroupOption[]> {
  const res = await api.get('/api/user/self/groups')
  const { data } = res

  if (!data.success || !data.data) {
    return []
  }

  const groupData = data.data as Record<string, { desc: string; ratio: number }>
  return Object.entries(groupData).map(([group, info]) => ({
    label: group,
    value: group,
    ratio: info.ratio,
    desc: info.desc,
  }))
}

export async function getRecentUserTasks(): Promise<unknown[]> {
  try {
    const res = await api.get('/api/task/self', {
      params: { p: 1, page_size: 20 },
      skipErrorHandler: true,
    } as Record<string, unknown>)
    if (res.data?.success && Array.isArray(res.data?.data?.items)) {
      return res.data.data.items
    }
    return []
  } catch {
    return []
  }
}
