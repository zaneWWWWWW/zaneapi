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
  readDataExportEnabled,
  readDataExportLastSuccessAt,
  readDataExportIntervalMinutes,
} from './freshness'

describe('quota data freshness', () => {
  test('reads the export interval from status payloads', () => {
    assert.equal(readDataExportIntervalMinutes({ data_export_interval: 5 }), 5)
    assert.equal(
      readDataExportIntervalMinutes({ data: { data_export_interval: '10' } }),
      10
    )
    assert.equal(readDataExportIntervalMinutes({ data_export_interval: 0 }), null)
    assert.equal(readDataExportIntervalMinutes(null), null)
  })

  test('reads the enabled flag and last successful export from status payloads', () => {
    assert.equal(readDataExportEnabled({ enable_data_export: false }), false)
    assert.equal(
      readDataExportLastSuccessAt({
        data: { data_export_last_success_at: '1234' },
      }),
      1234
    )
    assert.equal(readDataExportEnabled(null), null)
    assert.equal(readDataExportLastSuccessAt({}), null)
  })

  test('reports a disabled export instead of inventing a cutoff time', () => {
    const message = formatQuotaDataFreshnessMessage((key) => key, {
      status: { enable_data_export: false },
    })
    assert.equal(message, 'Dashboard data export is disabled.')
  })

  test('reports the actual last export time and interval', () => {
    const message = formatQuotaDataFreshnessMessage((key, options) => {
      if (
        key ===
        'Data last exported at {{time}}. Updated about every {{minutes}} minutes.'
      ) {
        return `exported ${options?.time}; every ${options?.minutes}`
      }
      return key
    }, {
      status: {
        enable_data_export: true,
        data_export_last_success_at: Date.parse('2026-03-12T14:37:22Z') / 1000,
        data_export_interval: 5,
      },
    })
    assert.match(message ?? '', /exported/)
    assert.match(message ?? '', /every 5/)
  })
})
