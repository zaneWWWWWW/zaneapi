import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import {
  buildStudioImageOptions,
  getStudioPriceDisplay,
  parseStudioModelKey,
  renderStudioPriceLabel,
  studioModelKey,
} from '../lib/studio-models'
import type { ModelOption } from '../types'

describe('studio combined model picker', () => {
  test('round-trips group and model in the select key', () => {
    const key = studioModelKey('vip', 'gpt-image-2')
    assert.equal(key, 'vip::gpt-image-2')
    assert.deepEqual(parseStudioModelKey(key), {
      group: 'vip',
      model: 'gpt-image-2',
    })
  })

  test('keeps the same model in different groups as separate options', () => {
    const models: ModelOption[] = [
      { label: 'gpt-image-2', value: 'gpt-image-2', type: 'image' },
    ]
    const options = buildStudioImageOptions({
      groups: [
        { label: 'default', value: 'default', ratio: 1, desc: '' },
        { label: 'vip', value: 'vip', ratio: 0.8, desc: '' },
      ],
      modelsByGroup: [
        { group: 'default', models },
        { group: 'vip', models },
      ],
      pricingModels: [
        {
          id: 1,
          model_name: 'gpt-image-2',
          quota_type: 1,
          model_ratio: 0,
          completion_ratio: 1,
          model_price: 0.04,
          enable_groups: ['default', 'vip'],
        },
      ],
      groupRatio: { default: 1, vip: 0.8 },
    })

    assert.equal(options.length, 2)
    assert.equal(options[0]?.group, 'default')
    assert.equal(options[1]?.group, 'vip')
    assert.equal(options[0]?.value, 'gpt-image-2')
    assert.equal(options[0]?.priceDisplay?.kind, 'request')
    assert.ok(
      renderStudioPriceLabel(options[0]?.priceDisplay, (key) => key).includes(
        'per request'
      )
    )
  })

  test('shows input and output prices for token-billed models', () => {
    const display = getStudioPriceDisplay(
      {
        id: 2,
        model_name: 'gemini-2.5-flash-image',
        quota_type: 0,
        model_ratio: 0.2,
        completion_ratio: 4,
        enable_groups: ['default'],
      },
      'default',
      { default: 1 }
    )
    assert.equal(display?.kind, 'token')
    if (display?.kind !== 'token') return
    assert.ok(display.input)
    assert.ok(display.output)
    assert.notEqual(display.input, display.output)
    const label = renderStudioPriceLabel(display, (key) => key)
    assert.ok(label.includes('Input'))
    assert.ok(label.includes('Output'))
    assert.ok(label.includes('/M'))
  })

  test('keeps only model names that contain image', () => {
    const options = buildStudioImageOptions({
      groups: [{ label: 'default', value: 'default', ratio: 1, desc: '' }],
      modelsByGroup: [
        {
          group: 'default',
          models: [
            { label: 'gpt-4o', value: 'gpt-4o', type: 'all' },
            { label: 'flux-1-dev', value: 'flux-1-dev', type: 'image' },
            { label: 'dall-e-3', value: 'dall-e-3', type: 'image' },
            { label: 'gpt-image-2', value: 'gpt-image-2', type: 'image' },
          ],
        },
      ],
      pricingModels: [],
      groupRatio: { default: 1 },
    })
    assert.deepEqual(
      options.map((model) => model.value),
      ['gpt-image-2']
    )
  })
})
