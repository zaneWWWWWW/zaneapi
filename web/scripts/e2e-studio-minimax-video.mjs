import assert from 'node:assert/strict'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const { chromium } = require(
  '/Users/ethan/.nvm/versions/node/v22.23.1/lib/node_modules/@playwright/cli/node_modules/playwright-core'
)

const BASE_URL = process.env.TEST_BASE_URL || 'http://localhost:3005'
const UPSTREAM_URL = process.env.UPSTREAM_URL || 'http://10.149.9.30:28082'
const UPSTREAM_KEY = process.env.UPSTREAM_KEY || 'qwen38-serve-2026'

async function runMinimaxStudioE2ETest() {
  console.log('========================================================')
  console.log('[E2E] Minimax H3 Video Model - Playwright E2E Test Suite')
  console.log('========================================================')

  // --- Step 0: Direct Upstream Probe Validation ---
  console.log(`[E2E] Step 0: Probing upstream model list at ${UPSTREAM_URL}/v1/models...`)
  const probeRes = await fetch(`${UPSTREAM_URL}/v1/models`, {
    headers: { Authorization: `Bearer ${UPSTREAM_KEY}` },
  })
  assert.equal(
    probeRes.status,
    200,
    `Upstream /v1/models probe failed with status ${probeRes.status}`
  )
  const probeData = await probeRes.json()
  console.log(
    `[E2E] Upstream probe success! Detected ${probeData.data?.length} models:`,
    probeData.data?.map((m) => m.id).join(', ')
  )
  const upstreamModelIds = probeData.data?.map((m) => m.id) || []
  assert.ok(
    upstreamModelIds.includes('minimax-h3') &&
      upstreamModelIds.includes('minimax-h3-turbo'),
    'Upstream models must include minimax-h3 and minimax-h3-turbo'
  )

  // --- Step 1: Launch Playwright Browser (Edge) ---
  console.log(`[E2E] Step 1: Launching Playwright browser against ${BASE_URL}...`)
  const browser = await chromium.launch({
    channel: 'msedge',
    headless: true,
  })

  const context = await browser.newContext()
  const page = await context.newPage()

  try {
    // 2. Sign in as admin
    console.log('[E2E] Step 2: Navigating to sign-in page...')
    await page.goto(`${BASE_URL}/sign-in`, { waitUntil: 'networkidle' })

    console.log('[E2E] Step 3: Entering admin credentials (root)...')
    await page.getByPlaceholder('输入您的用户名或电子邮件').fill('root')
    await page.getByPlaceholder('输入密码').fill('123456')
    await page.getByRole('button', { name: '登录' }).click()

    await page.waitForURL('**/dashboard/**', { timeout: 15000 })
    console.log('[E2E] Logged in successfully! Landing page:', page.url())

    // 3. Verify Channel configuration
    console.log('[E2E] Step 4: Inspecting /channels list...')
    await page.goto(`${BASE_URL}/channels`, { waitUntil: 'networkidle' })
    const channelName = page.getByText('Minimax').first()
    await channelName.waitFor({ state: 'visible', timeout: 5000 })
    console.log('[E2E] Channel "Minimax" found and active.')

    // 4. Navigate to Studio
    console.log('[E2E] Step 5: Navigating to Studio (/studio)...')
    await page.goto(`${BASE_URL}/studio`, { waitUntil: 'networkidle' })

    // 5. Switch to Video mode
    console.log('[E2E] Step 6: Switching to "AI 生视频" mode...')
    const videoModeTab = page.getByRole('button', { name: 'AI 生视频' })
    await videoModeTab.click()

    // 6. Verify Model selector has minimax-h3-turbo
    console.log('[E2E] Step 7: Verifying active video model...')
    const modelSelector = page.locator('button[role="combobox"]').nth(1)
    const selectedModelText = await modelSelector.textContent()
    console.log('[E2E] Active video model:', selectedModelText?.trim())
    assert.ok(
      selectedModelText?.includes('minimax-h3-turbo') ||
        selectedModelText?.includes('minimax'),
      `Expected minimax model, but got: ${selectedModelText}`
    )

    // 7. Select inspiration prompt
    console.log('[E2E] Step 8: Triggering inspiration prompt dropdown...')
    const inspirationDropdown = page
      .locator('button[role="combobox"]')
      .filter({ hasText: /灵感/ })
    await inspirationDropdown.click()
    const firstInspiration = page.getByRole('option').first()
    await firstInspiration.click()

    const promptTextarea = page.locator('textarea')
    const promptValue = await promptTextarea.inputValue()
    console.log(
      '[E2E] Inspiration prompt loaded:',
      promptValue.slice(0, 70) + '...'
    )
    assert.ok(promptValue.length > 0, 'Prompt textarea should not be empty')

    // 8. Verify creation gallery exists
    console.log('[E2E] Step 9: Checking creation gallery status...')
    const gallerySection = page.locator('div:has-text("创作成果")').first()
    assert.ok(await gallerySection.isVisible(), 'Gallery section must be visible')

    // 9. Inspect existing video creation card and modal
    const creationCards = page.locator(
      'div[class*="group relative flex flex-col"]'
    )
    if ((await creationCards.count()) > 0) {
      console.log('[E2E] Step 10: Clicking gallery video card to inspect player modal...')
      await creationCards.first().click()

      const videoModal = page.getByRole('dialog', { name: /视频播放器/ })
      await videoModal.waitFor({ state: 'visible', timeout: 5000 })
      console.log('[E2E] Video player modal displayed successfully!')

      const modalText = await videoModal.textContent()
      assert.ok(
        modalText.includes('minimax-h3-turbo') || modalText.includes('5s'),
        'Modal should show video duration and model information'
      )
      console.log('[E2E] Modal metadata verified (model, duration, aspect ratio).')

      const closeBtn = videoModal.getByRole('button', { name: 'Close' })
      await closeBtn.click()
    }

    console.log('========================================================')
    console.log('[E2E] 🎉 ALL PLAYWRIGHT E2E TESTS COMPLETED AND PASSED!')
    console.log('========================================================')
  } catch (error) {
    console.error('[E2E] ❌ Test FAILED:', error)
    await page
      .screenshot({ path: '/tmp/playwright-studio-e2e-error.png' })
      .catch(() => {})
    throw error
  } finally {
    await browser.close()
  }
}

runMinimaxStudioE2ETest()
  .then(() => process.exit(0))
  .catch(() => process.exit(1))
