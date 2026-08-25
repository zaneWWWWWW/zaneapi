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

import type { UptimeGroupResult } from '../../types'
import {
  buildProbeLookup,
  listUnmatchedMonitors,
  matchChannelProbe,
  summarizeProbes,
} from '../uptime-match'

const groups: UptimeGroupResult[] = [
  {
    categoryName: 'Core',
    monitors: [
      { name: 'OpenAI', uptime: 0.99, status: 1, group: 'API' },
      { name: 'claude-prod', uptime: 0.8, status: 0, group: 'API' },
    ],
  },
]

describe('uptime kuma channel matching', () => {
  test('matches a channel to a Kuma monitor by case-insensitive name', () => {
    const lookup = buildProbeLookup(groups)
    const matched = matchChannelProbe('openai', lookup)
    assert.equal(matched?.name, 'OpenAI')
    assert.equal(matchChannelProbe('missing', lookup), null)
  })

  test('lists probes that do not belong to any channel name', () => {
    const unmatched = listUnmatchedMonitors(groups, ['OpenAI'])
    assert.deepEqual(
      unmatched.map((item) => item.name),
      ['claude-prod']
    )
  })

  test('summarizes up and down probe counts', () => {
    const summary = summarizeProbes(groups[0]?.monitors ?? [])
    assert.deepEqual(summary, { up: 1, down: 1, other: 0, total: 2 })
  })
})
