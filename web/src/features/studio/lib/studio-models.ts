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

import { QUOTA_TYPE_VALUES } from '@/features/pricing/constants'
import { formatFixedPrice, formatGroupPrice } from '@/features/pricing/lib/price'
import type { PricingModel } from '@/features/pricing/types'

import type { GroupOption, ModelOption, StudioPriceDisplay } from '../types'

export function isStudioImageModelName(model: string): boolean {
  return model.toLowerCase().includes('image')
}

export function studioModelKey(group: string, model: string): string {
  return `${group}::${model}`
}

export function parseStudioModelKey(value: string): {
  group: string
  model: string
} {
  const separator = value.indexOf('::')
  if (separator < 0) {
    return { group: '', model: value }
  }
  return {
    group: value.slice(0, separator),
    model: value.slice(separator + 2),
  }
}

export function getStudioPriceDisplay(
  pricing: PricingModel | undefined,
  group: string,
  groupRatio: Record<string, number>
): StudioPriceDisplay | undefined {
  if (!pricing) return undefined
  if (pricing.quota_type === QUOTA_TYPE_VALUES.REQUEST) {
    const amount = formatFixedPrice(pricing, group, false, 1, 1, groupRatio)
    if (!amount || amount === '-') return undefined
    return { kind: 'request', amount }
  }

  const input = formatGroupPrice(
    pricing,
    group,
    'input',
    'M',
    false,
    1,
    1,
    groupRatio
  )
  const output = formatGroupPrice(
    pricing,
    group,
    'output',
    'M',
    false,
    1,
    1,
    groupRatio
  )
  const display: StudioPriceDisplay = { kind: 'token' }
  if (input && input !== '-') display.input = input
  if (output && output !== '-') display.output = output
  if (!display.input && !display.output) return undefined
  return display
}

export function renderStudioPriceLabel(
  display: StudioPriceDisplay | undefined,
  t: (key: string) => string
): string {
  if (!display) return ''
  if (display.kind === 'request') {
    return `${display.amount} / ${t('per request')}`
  }
  const parts: string[] = []
  if (display.input) {
    parts.push(`${t('Input')} ${display.input}/M`)
  }
  if (display.output) {
    parts.push(`${t('Output')} ${display.output}/M`)
  }
  return parts.join(' · ')
}

export function buildStudioImageOptions(input: {
  groups: GroupOption[]
  modelsByGroup: { group: string; models: ModelOption[] }[]
  pricingModels: PricingModel[]
  groupRatio: Record<string, number>
}): ModelOption[] {
  const pricingByName = new Map(
    input.pricingModels.map((model) => [model.model_name, model])
  )
  const groupMeta = new Map(input.groups.map((group) => [group.value, group]))
  const options: ModelOption[] = []

  for (const entry of input.modelsByGroup) {
    const group = groupMeta.get(entry.group)
    for (const model of entry.models) {
      if (!isStudioImageModelName(model.value)) continue
      const pricing = pricingByName.get(model.value)
      options.push({
        ...model,
        group: entry.group,
        groupRatio: group?.ratio,
        priceDisplay: getStudioPriceDisplay(
          pricing,
          entry.group,
          input.groupRatio
        ),
      })
    }
  }

  return options.sort((left, right) => {
    if (Boolean(left.isPopular) !== Boolean(right.isPopular)) {
      return left.isPopular ? -1 : 1
    }
    const byName = left.value.localeCompare(right.value)
    if (byName !== 0) return byName
    return (left.group || '').localeCompare(right.group || '')
  })
}
