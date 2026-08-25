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
  formatQuotaDataFreshnessMessage,
  getQuotaDataCutoff,
  readDataExportIntervalMinutes,
} from './freshness'

describe('quota data freshness', () => {
  test('floors the cutoff to the current hour', () => {
    const now = Date.parse('2026-03-12T14:37:22Z')
    const cutoff = getQuotaDataCutoff(now)
    assert.equal(cutoff.getMinutes(), 0)
    assert.equal(cutoff.getSeconds(), 0)
    assert.equal(cutoff.getMilliseconds(), 0)
    assert.ok(cutoff.getTime() <= now)
    assert.ok(now - cutoff.getTime() < 60 * 60 * 1000)
  })

  test('reads the export interval from status payloads', () => {
    assert.equal(readDataExportIntervalMinutes({ data_export_interval: 5 }), 5)
    assert.equal(
      readDataExportIntervalMinutes({ data: { data_export_interval: '10' } }),
      10
    )
    assert.equal(readDataExportIntervalMinutes({ data_export_interval: 0 }), null)
    assert.equal(readDataExportIntervalMinutes(null), null)
  })

  test('includes the refresh interval when it is configured', () => {
    const message = formatQuotaDataFreshnessMessage((key, options) => {
      if (
        key ===
        'Data through {{time}}. Updated about every {{minutes}} minutes.'
      ) {
        return `through ${options?.time}; every ${options?.minutes}`
      }
      return key
    }, {
      now: Date.parse('2026-03-12T14:37:22Z'),
      status: { data_export_interval: 5 },
    })
    assert.match(message, /every 5/)
  })
})
