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

import {
  blobFromDataUrl,
  buildImageRequestBody,
  collectStudioReferenceImages,
  creationReferenceSrc,
  referenceImageFilename,
} from '../lib/image-request'

// 线上可见的请求体是 JSON 序列化结果，undefined 字段不会被发送；
// structuredClone 会保留 undefined 字段，无法表达线上格式，这里必须走 JSON。
function wireBody(payload: Parameters<typeof buildImageRequestBody>[0]) {
  // eslint-disable-next-line unicorn/prefer-structured-clone
  return JSON.parse(JSON.stringify(buildImageRequestBody(payload))) as Record<
    string,
    unknown
  >
}

describe('studio image generation request body', () => {
  test('maps the standard quality level for gpt-image-2', () => {
    // Arrange & Act
    const body = wireBody({
      model: 'gpt-image-2',
      prompt: 'a red apple on a white background',
      quality: 'standard',
    })

    // Assert
    assert.equal(body.quality, 'medium')
  })

  test('maps hd and ultra quality levels for gpt-image-2 variants', () => {
    // Arrange & Act
    const hd = wireBody({ model: 'gpt-image-2.5-flare', prompt: 'p', quality: 'hd' })
    const ultra = wireBody({
      model: 'gpt-image-2.5-sunburst',
      prompt: 'p',
      quality: 'ultra',
    })

    // Assert
    assert.equal(hd.quality, 'high')
    assert.equal(ultra.quality, 'high')
  })

  test('omits style and response_format for gpt-image models', () => {
    // Arrange & Act
    const body = wireBody({
      model: 'gpt-image-2',
      prompt: 'p',
      quality: 'standard',
      style: 'vivid',
      response_format: 'url',
    })

    // Assert
    assert.equal('style' in body, false)
    assert.equal('response_format' in body, false)
  })

  test('keeps dall-e style parameters for models that support them', () => {
    // Arrange & Act
    const body = wireBody({
      model: 'dall-e-3',
      prompt: 'p',
      quality: 'hd',
      style: 'vivid',
      response_format: 'url',
    })

    // Assert
    assert.equal(body.quality, 'hd')
    assert.equal(body.style, 'vivid')
    assert.equal(body.response_format, 'url')
  })

  test('omits response_format for providers that do not accept it', () => {
    // Arrange & Act
    const body = wireBody({
      model: 'gemini-3.1-flash-image-preview-time',
      prompt: 'p',
      style: 'vivid',
    })

    // Assert
    assert.equal(body.style, 'vivid')
    assert.equal('response_format' in body, false)
  })

  test('applies the server-side defaults for count and size', () => {
    // Arrange & Act
    const body = wireBody({ model: 'gpt-image-2', prompt: 'p' })

    // Assert
    assert.equal(body.n, 1)
    assert.equal(body.size, '1024x1024')
  })

  test('prefers local base64 over a remote url for remix', () => {
    assert.equal(
      creationReferenceSrc({
        url: 'https://cdn.example/a.png',
        b64Json: 'abc',
      }),
      'data:image/png;base64,abc'
    )
  })

  test('decodes a data URL without fetch', () => {
    const blob = blobFromDataUrl('data:image/png;base64,AQID')
    assert.equal(blob.type, 'image/png')
    assert.equal(blob.size, 3)
  })

  test('names reference files from the blob mime type', () => {
    assert.equal(referenceImageFilename(0, 'image/jpeg'), 'reference-1.jpg')
    assert.equal(referenceImageFilename(1, 'image/webp'), 'reference-2.webp')
    assert.equal(referenceImageFilename(2, 'image/png'), 'reference-3.png')
  })

  test('collects unique data-URI reference images for edits', () => {
    const images = collectStudioReferenceImages({
      model: 'gpt-image-2',
      prompt: 'p',
      image: 'data:image/png;base64,aaa',
      images: [
        'data:image/png;base64,bbb',
        'https://example.com/skip.png',
        'data:image/png;base64,aaa',
      ],
    })
    assert.deepEqual(images, [
      'data:image/png;base64,bbb',
      'data:image/png;base64,aaa',
    ])
  })

  test('includes the selected playground group in the request body', () => {
    const body = wireBody({
      model: 'gpt-image-2.5-flare-times',
      prompt: 'p',
      group: 'vip',
    })
    assert.equal(body.group, 'vip')
  })
})
