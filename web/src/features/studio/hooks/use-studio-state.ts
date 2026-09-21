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

import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { getPerfMetricsSummary } from '@/features/performance-metrics/api'

import {
  generateImage,
  generateVideo,
  getUserAvailableModels,
  getUserGroups,
} from '../api'
import {
  DEFAULT_GROUP,
  IMAGE_RATIO_OPTIONS,
  POPULAR_IMAGE_MODELS,
  POPULAR_VIDEO_MODELS,
  STUDIO_STORAGE_KEY,
  STUDIO_VIDEO_WARNED_KEY,
} from '../constants'
import type {
  CameraMotion,
  CreationItem,
  GroupOption,
  ImageAspectRatio,
  ImageQuality,
  ImageStyle,
  ModelOption,
  StudioMode,
  VideoAspectRatio,
  VideoDuration,
  VideoResolution,
} from '../types'
import { reviveSavedCreations } from '../lib/creations-storage'
import { attachModelSuccessRates } from '../lib/model-success-rate'
import { useTaskPoller } from './use-task-poller'

function loadSavedCreations(): CreationItem[] {
  if (typeof window === 'undefined') return []
  try {
    const raw = window.localStorage.getItem(STUDIO_STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return reviveSavedCreations(parsed as CreationItem[])
  } catch {
    return []
  }
}

function saveCreations(items: CreationItem[]) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(
      STUDIO_STORAGE_KEY,
      JSON.stringify(items.slice(0, 100))
    )
  } catch {
    /* ignore storage quota errors */
  }
}

export function useStudioState(mode: StudioMode) {
  const { t } = useTranslation()
  const [prompt, setPrompt] = useState<string>('')
  const [negativePrompt, setNegativePrompt] = useState<string>('')
  const [selectedGroup, setSelectedGroup] = useState<string>(DEFAULT_GROUP)
  const [models, setModels] = useState<ModelOption[]>([])
  const [groups, setGroups] = useState<GroupOption[]>([])

  const [selectedImageModel, setSelectedImageModel] = useState<string>(
    POPULAR_IMAGE_MODELS[0]
  )
  const [selectedVideoModel, setSelectedVideoModel] = useState<string>(
    POPULAR_VIDEO_MODELS[0]
  )

  const [aspectRatio, setAspectRatio] = useState<ImageAspectRatio>('1:1')
  const [videoAspectRatio, setVideoAspectRatio] =
    useState<VideoAspectRatio>('16:9')
  const [videoDuration, setVideoDuration] = useState<VideoDuration>(5)
  const [videoResolution, setVideoResolution] =
    useState<VideoResolution>('720p')
  const [cameraMotion, setCameraMotion] = useState<CameraMotion>('auto')
  const [imageCount, setImageCount] = useState<number>(1)
  const [imageQuality, setImageQuality] = useState<ImageQuality>('standard')
  const [imageStyle, setImageStyle] = useState<ImageStyle>('vivid')
  const [referenceImage, setReferenceImage] = useState<string | null>(null)
  const [lastFrameImage, setLastFrameImage] = useState<string | null>(null)

  const [creations, setCreations] = useState<CreationItem[]>(() =>
    loadSavedCreations()
  )
  const [isGenerating, setIsGenerating] = useState<boolean>(false)

  // Save to local storage on creation list updates
  useEffect(() => {
    saveCreations(creations)
  }, [creations])

  // Update single creation by ID
  const updateCreation = useCallback(
    (id: string, patch: Partial<CreationItem>) => {
      setCreations((prev) =>
        prev.map((item) => (item.id === id ? { ...item, ...patch } : item))
      )
    },
    []
  )

  // Poller for in-progress video tasks
  useTaskPoller({ creations, updateCreation })

  // Fetch user groups on mount
  useEffect(() => {
    getUserGroups()
      .then((data) => {
        setGroups(data)
        if (data.length > 0 && !data.some((g) => g.value === selectedGroup)) {
          setSelectedGroup(data[0].value)
        }
      })
      .catch(() => {})
  }, [selectedGroup])

  // Fetch available models whenever selectedGroup changes
  useEffect(() => {
    let cancelled = false
    Promise.all([
      getUserAvailableModels(selectedGroup),
      getPerfMetricsSummary(24).catch(() => null),
    ])
      .then(([data, summary]) => {
        if (cancelled) return
        const withRates = attachModelSuccessRates(
          data,
          summary?.data.models ?? []
        )
        setModels(withRates)
        const imageMatches = withRates.filter(
          (m) => m.type === 'image' || m.type === 'all'
        )
        if (imageMatches.length > 0) {
          const preferred =
            imageMatches.find((m) =>
              (POPULAR_IMAGE_MODELS as readonly string[]).includes(m.value)
            ) || imageMatches[0]
          setSelectedImageModel(preferred.value)
        }
        const videoMatches = withRates.filter(
          (m) => m.type === 'video' || m.type === 'all'
        )
        if (videoMatches.length > 0) {
          const preferred =
            videoMatches.find((m) =>
              (POPULAR_VIDEO_MODELS as readonly string[]).includes(m.value)
            ) || videoMatches[0]
          setSelectedVideoModel(preferred.value)
        }
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [selectedGroup])

  const [videoNoticeOpen, setVideoNoticeOpen] = useState<boolean>(false)

  // Browser-level exit protection (beforeunload) while video generation is active
  const hasActiveVideoTasks = useMemo(
    () =>
      creations.some(
        (c) =>
          c.type === 'video' &&
          (c.status === 'queued' || c.status === 'in_progress')
      ),
    [creations]
  )

  useEffect(() => {
    if (!hasActiveVideoTasks) return

    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault()
      e.returnValue = ''
      return ''
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
  }, [hasActiveVideoTasks])

  // Filter models based on current active mode
  const currentAvailableModels = useMemo(() => {
    if (mode === 'image') {
      const filtered = models.filter((m) => m.type === 'image')
      return filtered.length > 0 ? filtered : models
    }
    const filtered = models.filter((m) => m.type === 'video')
    return filtered.length > 0 ? filtered : models
  }, [mode, models])

  // Video generation executor
  const executeVideoGenerate = useCallback(async () => {
    const trimmedPrompt = prompt.trim()
    if (!trimmedPrompt) return

    setIsGenerating(true)
    const selectedModel = selectedVideoModel
    const tempId = `vid-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`

    const newCreation: CreationItem = {
      id: tempId,
      type: 'video',
      status: 'queued',
      prompt: trimmedPrompt,
      model: selectedModel,
      createdAt: Date.now(),
      aspectRatio: videoAspectRatio,
      duration: videoDuration,
      progress: 0,
      group: selectedGroup,
    }

    setCreations((prev) => [newCreation, ...prev])

    try {
      const res = await generateVideo({
        model: selectedModel,
        prompt: trimmedPrompt,
        duration: videoDuration,
        aspect_ratio: videoAspectRatio,
        prompt_image: referenceImage || undefined,
        last_frame_image: lastFrameImage || undefined,
        camera_motion: cameraMotion !== 'auto' ? cameraMotion : undefined,
        resolution: videoResolution,
        group: selectedGroup,
      })

      if (res.error) {
        updateCreation(tempId, {
          status: 'failed',
          failReason: res.error.message || t('Video task submission failed'),
        })
        toast.error(res.error.message || t('Video task submission failed'))
        return
      }

      const taskId = res.id
      const anyRes = res as unknown as {
        status?: string
        url?: string
        data?: Array<{ url?: string }>
      }
      if (
        anyRes.status === 'completed' ||
        (Array.isArray(anyRes.data) && anyRes.data[0]?.url)
      ) {
        const videoUrl =
          anyRes.url ||
          anyRes.data?.[0]?.url ||
          `/v1/videos/${taskId}/content`
        updateCreation(tempId, {
          taskId,
          status: 'completed',
          progress: 100,
          url: videoUrl,
        })
        toast.success(t('Video generated successfully!'))
        return
      }

      updateCreation(tempId, {
        taskId,
        status: 'in_progress',
        progress: 10,
      })
      toast.success(
        t('Video generation task submitted! Please keep this page open.')
      )
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: {
          data?: {
            error?: { message?: string }
            message?: string
          }
        }
        message?: string
      }
      const message =
        axiosErr.response?.data?.error?.message ||
        axiosErr.response?.data?.message ||
        (err instanceof Error ? err.message : '') ||
        t('Video task submission failed')
      updateCreation(tempId, {
        status: 'failed',
        failReason: message,
      })
      toast.error(message)
    } finally {
      setIsGenerating(false)
    }
  }, [
    cameraMotion,
    lastFrameImage,
    prompt,
    referenceImage,
    selectedGroup,
    selectedVideoModel,
    t,
    updateCreation,
    videoAspectRatio,
    videoDuration,
    videoResolution,
  ])

  const confirmVideoNotice = useCallback(() => {
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STUDIO_VIDEO_WARNED_KEY, 'true')
    }
    void executeVideoGenerate()
  }, [executeVideoGenerate])

  // Execute Generation
  const handleGenerate = useCallback(async () => {
    const trimmedPrompt = prompt.trim()
    if (!trimmedPrompt) {
      toast.error(t('Please enter a prompt for generation'))
      return
    }

    if (mode === 'video') {
      const warned =
        typeof window !== 'undefined' &&
        window.localStorage.getItem(STUDIO_VIDEO_WARNED_KEY) === 'true'
      if (!warned) {
        setVideoNoticeOpen(true)
        return
      }
      void executeVideoGenerate()
      return
    }

    setIsGenerating(true)

    const selectedModel = selectedImageModel
    const currentRatio =
      IMAGE_RATIO_OPTIONS.find((r) => r.value === aspectRatio)?.size ||
      '1024x1024'

    const tempId = `img-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
    const newCreation: CreationItem = {
      id: tempId,
      type: 'image',
      status: 'in_progress',
      prompt: trimmedPrompt,
      model: selectedModel,
      createdAt: Date.now(),
      aspectRatio,
      group: selectedGroup,
    }

    setCreations((prev) => [newCreation, ...prev])

    try {
      const res = await generateImage({
        model: selectedModel,
        prompt: trimmedPrompt,
        negative_prompt: negativePrompt.trim() || undefined,
        size: currentRatio,
        n: imageCount,
        quality: imageQuality,
        style: imageStyle,
        image: referenceImage || undefined,
        group: selectedGroup,
      })

      if (res.error) {
        updateCreation(tempId, {
          status: 'failed',
          failReason: res.error.message || t('Image generation failed'),
        })
        toast.error(res.error.message || t('Image generation failed'))
        return
      }

      if (res.data && res.data.length > 0) {
        const firstImage = res.data[0]
        const imageUrl = firstImage.url
        const b64 = firstImage.b64_json

        updateCreation(tempId, {
          status: 'completed',
          progress: 100,
          url: imageUrl,
          b64Json: b64,
        })

        // If multiple images were requested and returned, add remaining
        if (res.data.length > 1) {
          const extraItems: CreationItem[] = res.data.slice(1).map((item, idx) => ({
            id: `${tempId}-${idx + 1}`,
            type: 'image',
            status: 'completed',
            prompt: trimmedPrompt,
            model: selectedModel,
            createdAt: Date.now(),
            aspectRatio,
            url: item.url,
            b64Json: item.b64_json,
            group: selectedGroup,
          }))
          setCreations((prev) => [...extraItems, ...prev])
        }

        toast.success(t('Image generated successfully!'), {
          description: t(
            'Generated images are stored only in this browser. Download them promptly after generation.'
          ),
        })
      } else {
        updateCreation(tempId, {
          status: 'failed',
          failReason: t('No image data returned from provider'),
        })
        toast.error(t('No image data returned from provider'))
      }
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: {
          data?: {
            error?: { message?: string }
            message?: string
          }
        }
        message?: string
      }
      const message =
        axiosErr.response?.data?.error?.message ||
        axiosErr.response?.data?.message ||
        (err instanceof Error ? err.message : '') ||
        t('Image generation failed')
      updateCreation(tempId, {
        status: 'failed',
        failReason: message,
      })
      toast.error(message)
    } finally {
      setIsGenerating(false)
    }
  }, [
    aspectRatio,
    executeVideoGenerate,
    imageCount,
    imageQuality,
    imageStyle,
    mode,
    negativePrompt,
    prompt,
    referenceImage,
    selectedGroup,
    selectedImageModel,
    t,
    updateCreation,
  ])

  const deleteCreation = useCallback((id: string) => {
    setCreations((prev) => prev.filter((item) => item.id !== id))
  }, [])

  const clearCreations = useCallback(() => {
    setCreations([])
    if (typeof window !== 'undefined') {
      window.localStorage.removeItem(STUDIO_STORAGE_KEY)
    }
  }, [])

  return {
    mode,
    prompt,
    setPrompt,
    negativePrompt,
    setNegativePrompt,
    selectedGroup,
    setSelectedGroup,
    groups,
    models,
    currentAvailableModels,
    selectedImageModel,
    setSelectedImageModel,
    selectedVideoModel,
    setSelectedVideoModel,
    aspectRatio,
    setAspectRatio,
    videoAspectRatio,
    setVideoAspectRatio,
    videoDuration,
    setVideoDuration,
    videoResolution,
    setVideoResolution,
    cameraMotion,
    setCameraMotion,
    imageCount,
    setImageCount,
    imageQuality,
    setImageQuality,
    imageStyle,
    setImageStyle,
    referenceImage,
    setReferenceImage,
    lastFrameImage,
    setLastFrameImage,
    creations,
    isGenerating,
    videoNoticeOpen,
    setVideoNoticeOpen,
    confirmVideoNotice,
    handleGenerate,
    deleteCreation,
    clearCreations,
  }
}
