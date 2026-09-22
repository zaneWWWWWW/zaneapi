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

import type {
  CameraMotion,
  ImageAspectRatio,
  InspirationPrompt,
  VideoAspectRatio,
  VideoDuration,
  VideoResolution,
} from './types'

export const STUDIO_STORAGE_KEY = 'zaneapi_studio_creations_v1'
export const STUDIO_VIDEO_WARNED_KEY = 'zaneapi_studio_video_warned_v1'
export const MAX_STUDIO_REFERENCE_IMAGES = 8
export const MAX_STUDIO_REFERENCE_IMAGE_BYTES = 15 * 1024 * 1024
export const MAX_STUDIO_IMAGE_N = 128
export const PRESET_STUDIO_IMAGE_COUNTS = [1, 2, 3, 4] as const

export const DEFAULT_GROUP = 'default' as const

export const POPULAR_IMAGE_MODELS = [
  'flux-1-schnell',
  'flux-1-dev',
  'flux-1-pro',
  'black-forest-labs/flux-1.1-pro',
  'dall-e-3',
  'agnes-image-2.5-flash',
  'stable-diffusion-3.5-large',
  'recraft-v3',
  'ideogram-v2',
] as const

export const POPULAR_VIDEO_MODELS = [
  'minimax-h3-turbo',
  'minimax-h3',
  'kling-v1',
  'kling-v1.5',
  'kling-v2',
  'sora-2',
  'grok-imagine-video',
  'agnes-video-2.5',
  'agnes-video-2.5-flash',
  'video-01',
  'cogvideox-flash',
] as const

export function isDalle3(model: string): boolean {
  return model.toLowerCase().includes('dall-e-3')
}

export function isFluxModel(model: string): boolean {
  return model.toLowerCase().includes('flux')
}

export function isSoraModel(model: string): boolean {
  return model.toLowerCase().includes('sora')
}

export function isKlingModel(model: string): boolean {
  return model.toLowerCase().includes('kling')
}

export interface ImageRatioOption {
  label: string
  value: ImageAspectRatio
  size: string
  descKey: string
  iconWidth: number
  iconHeight: number
}

export const IMAGE_RATIO_OPTIONS: ImageRatioOption[] = [
  {
    label: '1:1',
    value: '1:1',
    size: '1024x1024',
    descKey: 'Square',
    iconWidth: 18,
    iconHeight: 18,
  },
  {
    label: '16:9',
    value: '16:9',
    size: '1792x1024',
    descKey: 'Landscape',
    iconWidth: 24,
    iconHeight: 14,
  },
  {
    label: '9:16',
    value: '9:16',
    size: '1024x1792',
    descKey: 'Portrait',
    iconWidth: 14,
    iconHeight: 24,
  },
  {
    label: '4:3',
    value: '4:3',
    size: '1152x864',
    descKey: 'Classic 4:3',
    iconWidth: 20,
    iconHeight: 15,
  },
  {
    label: '3:4',
    value: '3:4',
    size: '864x1152',
    descKey: 'Classic 3:4',
    iconWidth: 15,
    iconHeight: 20,
  },
  {
    label: '3:2',
    value: '3:2',
    size: '1216x832',
    descKey: 'Photo 3:2',
    iconWidth: 21,
    iconHeight: 14,
  },
  {
    label: '21:9',
    value: '21:9',
    size: '1792x768',
    descKey: 'Cinematic',
    iconWidth: 26,
    iconHeight: 11,
  },
]

export interface VideoRatioOption {
  label: string
  value: VideoAspectRatio
  descKey: string
  iconWidth: number
  iconHeight: number
}

export const VIDEO_RATIO_OPTIONS: VideoRatioOption[] = [
  {
    label: '16:9',
    value: '16:9',
    descKey: 'Landscape (16:9)',
    iconWidth: 24,
    iconHeight: 14,
  },
  {
    label: '9:16',
    value: '9:16',
    descKey: 'Portrait (9:16)',
    iconWidth: 14,
    iconHeight: 24,
  },
  {
    label: '1:1',
    value: '1:1',
    descKey: 'Square (1:1)',
    iconWidth: 18,
    iconHeight: 18,
  },
  {
    label: '4:3',
    value: '4:3',
    descKey: 'Classic 4:3',
    iconWidth: 20,
    iconHeight: 15,
  },
  {
    label: '21:9',
    value: '21:9',
    descKey: 'Cinematic',
    iconWidth: 26,
    iconHeight: 11,
  },
]

export const VIDEO_DURATIONS: VideoDuration[] = [5, 10, 15]

export const VIDEO_RESOLUTIONS: Array<{ label: string; value: VideoResolution }> = [
  { label: '720p', value: '720p' },
  { label: '1080p', value: '1080p' },
]

export const CAMERA_MOTIONS: Array<{ labelKey: string; value: CameraMotion }> = [
  { labelKey: 'Auto Motion', value: 'auto' },
  { labelKey: 'Zoom In', value: 'zoom_in' },
  { labelKey: 'Zoom Out', value: 'zoom_out' },
  { labelKey: 'Pan Left', value: 'pan_left' },
  { labelKey: 'Pan Right', value: 'pan_right' },
]

export const IMAGE_INSPIRATIONS: InspirationPrompt[] = [
  {
    id: 'img-1',
    categoryKey: 'Sci-Fi & Cyberpunk',
    titleKey: 'Cyberpunk Metropolis',
    prompt:
      'A futuristic cyberpunk city with neon reflections in wet rain, hyper-detailed, 8k resolution, volumetric lighting',
  },
  {
    id: 'img-2',
    categoryKey: 'Portrait & Photography',
    titleKey: 'Cinematic Portrait',
    prompt:
      'Cinematic portrait of a courageous explorer in golden hour lighting, 35mm film photograph, highly detailed facial textures',
  },
  {
    id: 'img-3',
    categoryKey: '3D & Design',
    titleKey: 'Isometric Floating Island',
    prompt:
      'Minimalist 3D isometric floating island with a cozy coffee house, soft pastel colors, octane render, clean aesthetic',
  },
  {
    id: 'img-4',
    categoryKey: 'Nature & Macro',
    titleKey: 'Bioluminescent Forest',
    prompt:
      'Macro photography of glowing crystal mushrooms on a mystical mossy forest floor, soft ethereal rim lighting',
  },
  {
    id: 'img-5',
    categoryKey: 'Art & Illustration',
    titleKey: 'Ink Watercolor Dragon',
    prompt:
      'Dynamic watercolor splash of an ink-style dragon rising through stormy clouds, traditional East Asian calligraphy art',
  },
  {
    id: 'img-6',
    categoryKey: 'Architecture',
    titleKey: 'Modernist Villa',
    prompt:
      'Stunning contemporary villa cantilevered over an ocean cliff, floor-to-ceiling glass, warm interior ambient lights, sunset',
  },
]

export const VIDEO_INSPIRATIONS: InspirationPrompt[] = [
  {
    id: 'vid-1',
    categoryKey: 'Aerial & Cinematic',
    titleKey: 'FPV Drone Sci-Fi Flight',
    prompt:
      'FPV drone cinematic flythrough shot gliding through an ethereal futuristic sci-fi metropolis, high speed, volumetric lighting',
  },
  {
    id: 'vid-2',
    categoryKey: 'Action & Speed',
    titleKey: 'Neon City Drift',
    prompt:
      'A sleek sports car drifting around a wet corner in a neon city at night, realistic reflections, cinematic motion blur',
  },
  {
    id: 'vid-3',
    categoryKey: 'Nature & Atmosphere',
    titleKey: 'Sunset Wheat Field',
    prompt:
      'Gentle summer breeze blowing across a golden wheat field with warm sunset rim lighting, soft natural camera pan',
  },
  {
    id: 'vid-4',
    categoryKey: 'Time-lapse',
    titleKey: 'Aurora Peak Time-lapse',
    prompt:
      'Time-lapse of shimmering green aurora borealis dancing over a snowy jagged mountain peak under a starry sky',
  },
  {
    id: 'vid-5',
    categoryKey: 'Macro Motion',
    titleKey: 'Slow Motion Droplet',
    prompt:
      'Super slow motion water droplet splashing into calm crystal lake, morning golden sunlight refraction, ultra-crisp ripples',
  },
]
