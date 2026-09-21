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

import type { ImageGenerationRequest, ImageQuality } from '../types'

// gpt-image 系列（OpenAI 新版图片模型）的 quality 只接受 auto/low/medium/high，
// 没有 dall-e 的 style 参数，也不接受 response_format。studio 的三个档位按下表映射。
const GPT_IMAGE_QUALITY: Record<ImageQuality, string> = {
  standard: 'medium',
  hd: 'high',
  ultra: 'high',
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
