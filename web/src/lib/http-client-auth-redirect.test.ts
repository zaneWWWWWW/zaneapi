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
  isPublicPathAfterAuthLoss,
  resolveHttpErrorToastMessage,
} from './http-client'

describe('session-loss navigation', () => {
  test('keeps anonymous users on the homepage and other public pages', () => {
    assert.equal(isPublicPathAfterAuthLoss('/'), true)
    assert.equal(isPublicPathAfterAuthLoss('/sign-in'), true)
    assert.equal(isPublicPathAfterAuthLoss('/sign-up'), true)
    assert.equal(isPublicPathAfterAuthLoss('/oauth/github'), true)
    assert.equal(isPublicPathAfterAuthLoss('/privacy-policy'), true)
    assert.equal(isPublicPathAfterAuthLoss('/about'), true)
    assert.equal(isPublicPathAfterAuthLoss('/docs'), true)
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

describe('HTTP error toasts', () => {
  test('maps empty 429 responses to a rate-limit message', () => {
    assert.equal(
      resolveHttpErrorToastMessage({
        response: { status: 429, data: '' },
        message: 'Request failed with status code 429',
      }),
      'Too many requests'
    )
  })

  test('keeps business error text for non-429 responses', () => {
    assert.equal(
      resolveHttpErrorToastMessage({
        response: {
          status: 200,
          data: { success: false, message: '用户名或密码错误，或用户已被封禁' },
        },
      }),
      '用户名或密码错误，或用户已被封禁'
    )
  })
})
