import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { attachModelSuccessRates } from '../lib/model-success-rate'
import type { ModelOption } from '../types'

describe('attachModelSuccessRates', () => {
  test('copies 24h success rate onto matching studio models', () => {
    const models: ModelOption[] = [
      { label: 'gpt-image-2', value: 'gpt-image-2', type: 'image' },
      {
        label: 'gemini-3.1-flash-image-preview-time',
        value: 'gemini-3.1-flash-image-preview-time',
        type: 'image',
      },
    ]

    const attached = attachModelSuccessRates(models, [
      { model_name: 'GPT-image-2', success_rate: 97.5 },
      {
        model_name: 'gemini-3.1-flash-image-preview-time',
        success_rate: 81.25,
      },
    ])

    assert.equal(attached[0]?.successRate, 97.5)
    assert.equal(attached[1]?.successRate, 81.25)
  })

  test('leaves unmatched models without a rate', () => {
    const attached = attachModelSuccessRates(
      [{ label: 'flux-1', value: 'flux-1', type: 'image' }],
      [{ model_name: 'gpt-image-2', success_rate: 99 }]
    )
    assert.equal(attached[0]?.successRate, undefined)
  })
})
