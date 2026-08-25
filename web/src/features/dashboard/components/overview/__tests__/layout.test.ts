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

import { getOverviewContentLayout } from '../overview-content-layout'

const hidden = {
  isAdmin: false,
  showApiInfo: false,
  showAnnouncements: false,
  showFaq: false,
  showGroupAvailability: false,
}

describe('overview content grid', () => {
  test('hides the content grid when every panel is off', () => {
    const layout = getOverviewContentLayout(hidden)
    assert.equal(layout.show, false)
    assert.equal(layout.showLeft, false)
  })

  test('keeps a single left panel in one column instead of a leftover two-column cell', () => {
    const layout = getOverviewContentLayout({
      ...hidden,
      showFaq: true,
    })
    assert.equal(layout.pairApiAndAnnouncements, false)
    assert.equal(layout.leftClassName.includes('flex-col'), true)
    assert.equal(layout.leftClassName.includes('grid-cols-2'), false)
    assert.equal(layout.outerClassName.includes('xl:grid-cols-['), false)
  })

  test('does not pair FAQ with API info when announcements are hidden', () => {
    const layout = getOverviewContentLayout({
      ...hidden,
      showApiInfo: true,
      showFaq: true,
    })
    assert.equal(layout.pairApiAndAnnouncements, false)
    assert.equal(layout.showFaq, true)
    assert.equal(layout.showApiInfo, true)
  })

  test('pairs announcements and API info with announcements in the wider first column', () => {
    const layout = getOverviewContentLayout({
      ...hidden,
      showApiInfo: true,
      showAnnouncements: true,
    })
    assert.equal(layout.pairApiAndAnnouncements, true)
    assert.equal(
      layout.apiAnnouncementsRowClassName.includes(
        'lg:grid-cols-[minmax(0,1.25fr)_minmax(0,1fr)]'
      ),
      true
    )
  })

  test('stretches group availability beside left content and uses a fixed height when it stands alone', () => {
    const withLeft = getOverviewContentLayout({
      ...hidden,
      showAnnouncements: true,
      showGroupAvailability: true,
    })
    assert.equal(withLeft.groupAvailabilityFill, true)
    assert.equal(withLeft.outerClassName.includes('xl:items-stretch'), true)
    assert.equal(
      withLeft.groupAvailabilityItemClassName.includes('xl:h-full'),
      true
    )
    assert.equal(
      withLeft.groupAvailabilityItemClassName.includes('min-h-80'),
      true
    )
    assert.equal(withLeft.leftClassName.includes('xl:self-start'), true)

    const availabilityOnly = getOverviewContentLayout({
      ...hidden,
      showGroupAvailability: true,
    })
    assert.equal(availabilityOnly.groupAvailabilityFill, false)
    assert.equal(
      availabilityOnly.outerClassName.includes('xl:grid-cols-['),
      false
    )
    assert.equal(
      availabilityOnly.groupAvailabilityItemClassName.includes('xl:h-full'),
      false
    )
  })

  test('counts the admin performance panel as left content so group availability can fill', () => {
    const layout = getOverviewContentLayout({
      ...hidden,
      isAdmin: true,
      showGroupAvailability: true,
    })
    assert.equal(layout.showLeft, true)
    assert.equal(layout.showPerformance, true)
    assert.equal(layout.groupAvailabilityFill, true)
  })
})
