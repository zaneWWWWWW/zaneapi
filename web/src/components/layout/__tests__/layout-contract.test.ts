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
  APP_HEADER_CHROME,
  APP_SIDEBAR_METRICS,
  DEFAULT_APP_LAYOUT,
  resolveAppLayoutVariant,
} from '../constants'

describe('ChatGPT-style application shell', () => {
  test('uses a connected sidebar instead of a floating panel by default', () => {
    assert.equal(DEFAULT_APP_LAYOUT.variant, 'sidebar')
    assert.equal(DEFAULT_APP_LAYOUT.collapsible, 'icon')
  })

  test('keeps the desktop workspace rail stable across expanded and collapsed states', () => {
    assert.deepEqual(APP_SIDEBAR_METRICS, {
      desktopWidth: '16.25rem',
      mobileWidth: '17rem',
      collapsedWidth: '3.25rem',
    })
  })

  test('falls back to the connected sidebar when the layout cookie is missing or unknown', () => {
    assert.equal(resolveAppLayoutVariant(undefined), 'sidebar')
    assert.equal(resolveAppLayoutVariant(null), 'sidebar')
    assert.equal(resolveAppLayoutVariant(''), 'sidebar')
    assert.equal(resolveAppLayoutVariant('legacy-inset'), 'sidebar')
  })

  test('still honors an explicit workspace variant chosen in settings', () => {
    assert.equal(resolveAppLayoutVariant('sidebar'), 'sidebar')
    assert.equal(resolveAppLayoutVariant('inset'), 'inset')
    assert.equal(resolveAppLayoutVariant('floating'), 'floating')
  })

  test('keeps the top bar as a utility strip instead of a sitemap', () => {
    assert.equal(APP_HEADER_CHROME.workspaceNavLinks, false)
    assert.equal(APP_HEADER_CHROME.publicNavLinks, false)
    assert.equal(APP_HEADER_CHROME.workspaceSearch, 'none')
  })
})
