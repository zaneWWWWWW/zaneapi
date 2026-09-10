import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import type { PricingModel } from '../../types'
import { getDisplayGroup } from '../model-helpers'

const model: PricingModel = {
  id: 1,
  model_name: 'gpt-test',
  quota_type: 0,
  model_ratio: 1,
  enable_groups: ['gptPLUS', 'gptPRO'],
  group_ratio: { gptPLUS: 0.15, gptPRO: 0.2 },
}

describe('model card display group', () => {
  test('follows the selected group used for price calculation', () => {
    assert.equal(getDisplayGroup(model, 'gptPRO'), 'gptPRO')
  })

  test('falls back to the first enabled group when no group is selected', () => {
    assert.equal(getDisplayGroup(model, 'all'), 'gptPLUS')
  })

  test('falls back when the selected group is unavailable for the model', () => {
    assert.equal(getDisplayGroup(model, 'default'), 'gptPLUS')
  })
})
