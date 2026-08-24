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
  filterVisibleSyncChannels,
  hideSyncChannelId,
  parseHiddenSyncChannelIds,
  serializeHiddenSyncChannelIds,
} from '../hidden-sync-channels'

describe('hidden sync channels', () => {
  test('parses integer ids and ignores invalid entries', () => {
    assert.deepEqual(parseHiddenSyncChannelIds(null), [])
    assert.deepEqual(parseHiddenSyncChannelIds('not-json'), [])
    assert.deepEqual(
      parseHiddenSyncChannelIds('[-100,-101,-100,"x"]'),
      [-100, -101]
    )
  })

  test('hides selected official sources without deleting other rows', () => {
    const channels = [
      { id: -100, name: '官方倍率预设' },
      { id: -101, name: 'models.dev 价格预设' },
      { id: -102, name: 'PriceApple' },
    ]
    const hidden = hideSyncChannelId(hideSyncChannelId([], -100), -101)
    assert.deepEqual(
      filterVisibleSyncChannels(channels, hidden).map((channel) => channel.id),
      [-102]
    )
    assert.equal(serializeHiddenSyncChannelIds(hidden), '[-100,-101]')
  })

  test('restoring uses an empty hidden list and shows every source again', () => {
    const channels = [{ id: -100 }, { id: -102 }]
    assert.deepEqual(
      filterVisibleSyncChannels(channels, []).map((channel) => channel.id),
      [-100, -102]
    )
  })
})
