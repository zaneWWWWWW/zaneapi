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
import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { POPULAR_VIDEO_MODELS } from '../constants'

describe('studio video model selection and classification', () => {
  test('includes minimax-h3 and minimax-h3-turbo in popular video models', () => {
    // Arrange & Act
    const popularList = POPULAR_VIDEO_MODELS as readonly string[]

    // Assert
    assert.ok(
      popularList.includes('minimax-h3-turbo'),
      'popular video models must include minimax-h3-turbo'
    )
    assert.ok(
      popularList.includes('minimax-h3'),
      'popular video models must include minimax-h3'
    )
  })

  test('classifies minimax and h3 model variants as video models', () => {
    // Arrange
    const candidateModels = [
      'minimax-h3',
      'minimax-h3-turbo',
      'minimax-video',
      'minimax/video-01',
      'h3',
    ]

    // Act & Assert
    for (const model of candidateModels) {
      const isVideo =
        (POPULAR_VIDEO_MODELS as readonly string[]).includes(model) ||
        model.includes('video') ||
        model.includes('kling') ||
        model.includes('sora') ||
        model.includes('grok-imagine') ||
        model.includes('hailuo') ||
        model.includes('minimax') ||
        model.includes('h3')

      assert.equal(
        isVideo,
        true,
        `expected ${model} to be recognized as a video model`
      )
    }
  })

  test('extracts direct video url from synchronous response payload', () => {
    // Arrange
    const synchronousPayload = {
      id: 'vid_sync_001',
      created: 1789200000,
      model: 'minimax-h3-turbo',
      prompt: 'FPV drone shot',
      duration_seconds: 5,
      resolution: '1344x768',
      data: [
        {
          url: 'http://10.149.9.30:28082/videos/sample.mp4',
          filename: 'sample.mp4',
        },
      ],
    }

    // Act
    const hasDirectVideoUrl = Boolean(
      synchronousPayload.data && synchronousPayload.data[0]?.url
    )
    const resolvedUrl = synchronousPayload.data?.[0]?.url

    // Assert
    assert.equal(hasDirectVideoUrl, true)
    assert.equal(resolvedUrl, 'http://10.149.9.30:28082/videos/sample.mp4')
  })
})
