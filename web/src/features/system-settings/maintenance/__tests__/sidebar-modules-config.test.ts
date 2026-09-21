import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import {
  migrateChatStudioModules,
  parseSidebarModulesAdmin,
} from '../config'

describe('sidebar studio module split', () => {
  test('copies a legacy studio flag onto both image and video modules', () => {
    const migrated = migrateChatStudioModules({
      enabled: true,
      studio: false,
      playground: true,
    })

    assert.equal(migrated['studio-image'], false)
    assert.equal(migrated['studio-video'], false)
    assert.equal('studio' in migrated, false)
    assert.equal(migrated.playground, true)
  })

  test('does not overwrite studio-image or studio-video once they exist', () => {
    const migrated = migrateChatStudioModules({
      enabled: true,
      studio: false,
      'studio-image': true,
      'studio-video': false,
    })

    assert.equal(migrated['studio-image'], true)
    assert.equal(migrated['studio-video'], false)
    assert.equal('studio' in migrated, false)
  })

  test('parseSidebarModulesAdmin migrates stored studio=false instead of defaulting both on', () => {
    const parsed = parseSidebarModulesAdmin(
      JSON.stringify({
        chat: {
          enabled: true,
          studio: false,
          playground: true,
          chat: true,
        },
      })
    )

    assert.equal(parsed.chat['studio-image'], false)
    assert.equal(parsed.chat['studio-video'], false)
    assert.equal('studio' in parsed.chat, false)
  })
})
