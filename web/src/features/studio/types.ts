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

export type StudioMode = 'image' | 'video'

export type ImageAspectRatio =
  | '1:1'
  | '16:9'
  | '9:16'
  | '4:3'
  | '3:4'
  | '3:2'
  | '21:9'

export type VideoAspectRatio = '16:9' | '9:16' | '1:1' | '4:3' | '21:9'

export type VideoDuration = 5 | 10 | 15

export type VideoResolution = '720p' | '1080p'

export type CameraMotion =
  | 'auto'
  | 'zoom_in'
  | 'zoom_out'
  | 'pan_left'
  | 'pan_right'

export type ImageQuality = 'standard' | 'hd' | 'ultra'

export type ImageStyle = 'vivid' | 'natural' | 'anime'

export type CreationType = 'image' | 'video'

export type CreationStatus = 'queued' | 'in_progress' | 'completed' | 'failed'

export interface InspirationPrompt {
  id: string
  titleKey: string
  prompt: string
  categoryKey: string
}

export interface CreationItem {
  id: string
  type: CreationType
  status: CreationStatus
  prompt: string
  model: string
  createdAt: number
  aspectRatio?: string
  duration?: number
  progress?: number // 0 - 100
  url?: string
  b64Json?: string
  failReason?: string
  taskId?: string
  group?: string
}

export interface ImageGenerationRequest {
  model: string
  prompt: string
  negative_prompt?: string
  n?: number
  size?: string
  quality?: ImageQuality
  style?: ImageStyle
  response_format?: 'url' | 'b64_json'
  image?: string // Base64 data or data URL for image2image/edits
  group?: string
}

export interface VideoGenerationRequest {
  model: string
  prompt: string
  duration?: number
  aspect_ratio?: string
  prompt_image?: string // First-frame / reference image
  image?: string
  last_frame_image?: string
  camera_motion?: string
  resolution?: string
  group?: string
}

export interface OpenAIImageResponseItem {
  url?: string
  b64_json?: string
  revised_prompt?: string
}

export interface OpenAIImageResponse {
  created?: number
  data?: OpenAIImageResponseItem[]
  error?: {
    message?: string
    type?: string
    code?: string
  }
}

export interface OpenAIVideoResponse {
  id: string
  object?: string
  created_at?: number
  status: string
  progress?: number
  model?: string
  prompt?: string
  url?: string
  error?: {
    message?: string
    code?: string
  }
}

export interface ModelOption {
  label: string
  value: string
  isPopular?: boolean
  type?: 'image' | 'video' | 'all'
}

export interface GroupOption {
  label: string
  value: string
  ratio: number
  desc: string
}
