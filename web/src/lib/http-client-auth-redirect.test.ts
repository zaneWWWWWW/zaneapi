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

import { isPublicPathAfterAuthLoss } from './http-client'

describe('session-loss navigation', () => {
  test('keeps anonymous users on the homepage and other public pages', () => {
    assert.equal(isPublicPathAfterAuthLoss('/'), true)
    assert.equal(isPublicPathAfterAuthLoss('/sign-in'), true)
    assert.equal(isPublicPathAfterAuthLoss('/sign-up'), true)
    assert.equal(isPublicPathAfterAuthLoss('/oauth/github'), true)
    assert.equal(isPublicPathAfterAuthLoss('/privacy-policy'), true)
    assert.equal(isPublicPathAfterAuthLoss('/pricing'), true)
    assert.equal(isPublicPathAfterAuthLoss('/rankings'), true)
  })

  test('still treats workspace pages as protected after a lost session', () => {
    assert.equal(isPublicPathAfterAuthLoss('/pricing/gpt-4'), true)
    assert.equal(isPublicPathAfterAuthLoss('/dashboard'), false)
    assert.equal(isPublicPathAfterAuthLoss('/keys'), false)
    assert.equal(isPublicPathAfterAuthLoss('/system-settings/site/system-info'), false)
  })
})
