import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import {
  INTERRUPTED_GENERATION_MESSAGE,
  reviveSavedCreations,
} from '../lib/creations-storage'
import type { CreationItem } from '../types'

function item(patch: Partial<CreationItem>): CreationItem {
  return {
    id: 'img-1',
    type: 'image',
    status: 'completed',
    prompt: 'a cat',
    model: 'gpt-image-2',
    createdAt: 1,
    ...patch,
  }
}

describe('reviveSavedCreations', () => {
  test('marks in-progress image jobs as failed after reload', () => {
    const revived = reviveSavedCreations([
      item({ id: 'pending', status: 'in_progress' }),
    ])
    assert.equal(revived[0]?.status, 'failed')
    assert.equal(revived[0]?.failReason, INTERRUPTED_GENERATION_MESSAGE)
  })

  test('keeps pending videos that still have a task id', () => {
    const revived = reviveSavedCreations([
      item({
        id: 'vid-1',
        type: 'video',
        status: 'in_progress',
        taskId: 'task-1',
      }),
    ])
    assert.equal(revived[0]?.status, 'in_progress')
    assert.equal(revived[0]?.taskId, 'task-1')
  })

  test('marks completed images without url or base64 as failed', () => {
    const revived = reviveSavedCreations([item({ status: 'completed' })])
    assert.equal(revived[0]?.status, 'failed')
  })

  test('keeps completed images that still have a preview', () => {
    const revived = reviveSavedCreations([
      item({ status: 'completed', url: 'https://example.com/a.png' }),
    ])
    assert.equal(revived[0]?.status, 'completed')
  })
})
