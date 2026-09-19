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

import { buildImageRequestBody } from '../lib/image-request'

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
})
