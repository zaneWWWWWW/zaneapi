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
  bucketSecondsToMinutes,
  filterAvailabilityGroups,
  isHourAvailabilityWindow,
  parseGroupAvailabilityGroups,
} from '../group-availability'

describe('admin group availability selection', () => {
  test('treats a missing or empty admin list as no groups to show', () => {
    assert.deepEqual(parseGroupAvailabilityGroups(undefined), [])
    assert.deepEqual(parseGroupAvailabilityGroups('[]'), [])
    assert.deepEqual(parseGroupAvailabilityGroups([]), [])
  })

  test('reads the admin-selected groups from status or option JSON', () => {
    assert.deepEqual(parseGroupAvailabilityGroups(['vip', 'default']), [
      'vip',
      'default',
    ])
    assert.deepEqual(
      parseGroupAvailabilityGroups(JSON.stringify(['default', 'auto'])),
      ['default', 'auto']
    )
  })

  test('shows only the groups the admin selected', () => {
    const groups = [{ group: 'default' }, { group: 'vip' }, { group: 'auto' }]
    assert.deepEqual(filterAvailabilityGroups(groups, []), [])
    assert.deepEqual(filterAvailabilityGroups(groups, ['vip']), [
      { group: 'vip' },
    ])
  })
})

describe('group availability current bucket', () => {
  test('converts the metrics bucket length to minutes', () => {
    assert.equal(bucketSecondsToMinutes(60), 1)
    assert.equal(bucketSecondsToMinutes(300), 5)
    assert.equal(bucketSecondsToMinutes(3600), 60)
  })

  test('falls back to a 60-minute bucket when the interval is missing', () => {
    assert.equal(bucketSecondsToMinutes(null), 60)
    assert.equal(bucketSecondsToMinutes(0), 60)
    assert.equal(bucketSecondsToMinutes(Number.NaN), 60)
  })

  test('treats the default hour bucket as a 1h current window', () => {
    assert.equal(isHourAvailabilityWindow(60), true)
    assert.equal(isHourAvailabilityWindow(5), false)
    assert.equal(isHourAvailabilityWindow(1), false)
  })
})
