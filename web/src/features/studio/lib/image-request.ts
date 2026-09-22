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

import { api } from '@/lib/http-client'

import type {
  ImageGenerationRequest,
  ImageQuality,
  StudioReferenceImage,
} from '../types'

// gpt-image 系列（OpenAI 新版图片模型）的 quality 只接受 auto/low/medium/high，
// 没有 dall-e 的 style 参数，也不接受 response_format。studio 的三个档位按下表映射。
const GPT_IMAGE_QUALITY: Record<ImageQuality, string> = {
  standard: 'medium',
  hd: 'high',
  ultra: 'high',
}

export function referenceImageFilename(index: number, mimeType: string): string {
  let ext = 'png'
  if (mimeType === 'image/jpeg') {
    ext = 'jpg'
  } else if (mimeType === 'image/webp') {
    ext = 'webp'
  } else if (mimeType === 'image/gif') {
    ext = 'gif'
  }
  return `reference-${index + 1}.${ext}`
}

export function creationReferenceSrc(item: {
  url?: string
  b64Json?: string
}): string {
  if (item.b64Json) {
    return item.b64Json.startsWith('data:')
      ? item.b64Json
      : `data:image/png;base64,${item.b64Json}`
  }
  return item.url || ''
}

export function blobFromDataUrl(src: string): Blob {
  const comma = src.indexOf(',')
  if (!src.startsWith('data:') || comma < 0) {
    throw new Error('invalid data URL')
  }
  const header = src.slice(0, comma)
  const data = src.slice(comma + 1).replaceAll(/\s/g, '')
  const mime = /data:([^;,]+)/.exec(header)?.[1] || 'image/png'
  const binary = atob(data)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return new Blob([bytes], { type: mime })
}

export async function studioReferenceImageFromSrc(
  src: string,
  filename = 'reference.png'
): Promise<StudioReferenceImage> {
  let blob: Blob
  if (src.startsWith('data:')) {
    blob = blobFromDataUrl(src)
  } else if (src.startsWith('blob:')) {
    const response = await fetch(src)
    if (!response.ok) {
      throw new Error('failed to read reference image')
    }
    blob = await response.blob()
  } else {
    const res = await api.get('/pg/images/file', {
      params: { url: src },
      responseType: 'blob',
      skipErrorHandler: true,
    })
    blob = res.data as Blob
    if (!(blob instanceof Blob) || blob.size === 0) {
      throw new Error('failed to read reference image')
    }
    if (blob.type.includes('json') || blob.type.startsWith('text/')) {
      throw new Error('failed to read reference image')
    }
  }
  const type = blob.type || 'image/png'
  const file = new File([blob], filename, { type })
  return {
    id: `${file.name}-${file.size}-${Date.now()}-${Math.random().toString(36).slice(2)}`,
    preview: URL.createObjectURL(file),
    file,
  }
}

export function triggerBlobDownload(blob: Blob, filename: string) {
  const objectUrl = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = objectUrl
  link.download = filename
  link.rel = 'noopener'
  link.style.display = 'none'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.setTimeout(() => URL.revokeObjectURL(objectUrl), 2000)
}

export async function downloadStudioImage(item: {
  id: string
  url?: string
  b64Json?: string
}) {
  const src = creationReferenceSrc(item)
  if (!src) {
    throw new Error('no image')
  }
  const ready = await studioReferenceImageFromSrc(src, `image-${item.id}.png`)
  if (!ready.file) {
    throw new Error('no image')
  }
  let ext = 'png'
  if (ready.file.type.includes('jpeg')) {
    ext = 'jpg'
  } else if (ready.file.type.includes('webp')) {
    ext = 'webp'
  }
  triggerBlobDownload(ready.file, `image-${item.id}.${ext}`)
  URL.revokeObjectURL(ready.preview)
}

export function collectStudioReferenceImages(
  payload: ImageGenerationRequest
): string[] {
  const images = [...(payload.images ?? []), ...(payload.image ? [payload.image] : [])]
  return images.filter(
    (src, index, all) => src.startsWith('data:') && all.indexOf(src) === index
  )
}

export function buildImageRequestBody(
  payload: ImageGenerationRequest
): Record<string, unknown> {
  const modelKey = payload.model.trim().toLowerCase()
  const isGptImage =
    modelKey.startsWith('gpt-image') || modelKey.startsWith('chatgpt-image')

  return {
    model: payload.model,
    prompt: payload.prompt,
    group: payload.group,
    n: payload.n || 1,
    size: payload.size || '1024x1024',
    // gpt-image 系列使用自己的参数面：quality 取值不同，也没有 style 与 response_format。
    quality: isGptImage
      ? GPT_IMAGE_QUALITY[payload.quality ?? 'standard']
      : payload.quality,
    style: isGptImage ? undefined : payload.style,
    // 其余模型按需透传，未显式要求时由服务端下发默认值。
    response_format: isGptImage ? undefined : payload.response_format,
  }
}
