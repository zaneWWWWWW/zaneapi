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

import type { PricingModel } from '../../types'
import {
  formatPrice,
  formatSquareGroupRatio,
  getSquareRequestPricePair,
  getSquareTokenPricePair,
} from '../price'

function tokenModel(overrides: Partial<PricingModel> = {}): PricingModel {
  return {
    id: 1,
    model_name: 'gpt-4.1-mini',
    quota_type: 0,
    model_ratio: 0.2,
    completion_ratio: 4,
    enable_groups: ['default', 'vip'],
    group_ratio: { default: 1, vip: 0.5 },
    ...overrides,
  }
}

describe('model square price pairs', () => {
  test('shows the model price and the cheapest usable group price when no group is selected', () => {
    const pair = getSquareTokenPricePair(tokenModel(), 'input', 'M')

    assert.equal(pair.groupRatio, 0.5)
    assert.equal(pair.differs, true)
    assert.notEqual(pair.own, pair.grouped)
    assert.equal(formatPrice(tokenModel(), 'input', 'M'), pair.grouped)
  })

  test('uses the selected group ratio for the grouped price and keeps the model price unchanged', () => {
    const model = tokenModel()
    const vip = getSquareTokenPricePair(model, 'input', 'M', false, 1, 1, 'vip')
    const def = getSquareTokenPricePair(
      model,
      'input',
      'M',
      false,
      1,
      1,
      'default'
    )

    assert.equal(vip.own, def.own)
    assert.equal(vip.groupRatio, 0.5)
    assert.equal(vip.differs, true)
    assert.equal(def.groupRatio, 1)
    assert.equal(def.differs, false)
    assert.equal(def.own, def.grouped)
  })

  test('request models keep the configured price as own and multiply it by group ratio', () => {
    const model = tokenModel({
      quota_type: 1,
      model_price: 0.5,
      model_ratio: 0,
      completion_ratio: 1,
    })
    const pair = getSquareRequestPricePair(model)

    assert.equal(pair.groupRatio, 0.5)
    assert.equal(pair.differs, true)
    assert.notEqual(pair.own, pair.grouped)
  })

  test('formats compact group ratio labels without trailing zeros', () => {
    assert.equal(formatSquareGroupRatio(1), '1')
    assert.equal(formatSquareGroupRatio(0.5), '0.5')
    assert.equal(formatSquareGroupRatio(1.25), '1.25')
  })
})
