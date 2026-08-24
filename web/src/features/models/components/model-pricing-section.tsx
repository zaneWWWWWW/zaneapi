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
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { SettingsPageProvider } from '@/features/system-settings/components/settings-page-context'
import {
  getOptionValue,
  useSystemOptions,
} from '@/features/system-settings/hooks/use-system-options'
import {
  RatioSettingsCard,
  type RatioTabId,
} from '@/features/system-settings/models/ratio-settings-card'

const PRICING_OPTION_DEFAULTS = {
  ModelPrice: '',
  ModelRatio: '',
  CacheRatio: '',
  CreateCacheRatio: '',
  CompletionRatio: '',
  ImageRatio: '',
  AudioRatio: '',
  AudioCompletionRatio: '',
  ExposeRatioEnabled: false,
  'billing_setting.billing_mode': '{}',
  'billing_setting.billing_expr': '{}',
  'tool_price_setting.prices': '{}',
  TopupGroupRatio: '',
  GroupRatio: '',
  UserUsableGroups: '',
  GroupGroupRatio: '',
  AutoGroups: '',
  DefaultUseAutoGroup: false,
  'group_ratio_setting.group_special_usable_group': '{}',
}

const DEFAULT_PRICING_TABS: RatioTabId[] = [
  'models',
  'unset-models',
  'tool-prices',
  'upstream-sync',
]

export function ModelPricingSection({
  visibleTabs = DEFAULT_PRICING_TABS,
}: {
  visibleTabs?: RatioTabId[]
}) {
  const { t } = useTranslation()
  const { data, isLoading } = useSystemOptions()
  const [titleStatusContainer, setTitleStatusContainer] =
    useState<HTMLSpanElement | null>(null)
  const [actionsContainer, setActionsContainer] =
    useState<HTMLDivElement | null>(null)

  const settings = useMemo(
    () => getOptionValue(data?.data, PRICING_OPTION_DEFAULTS),
    [data?.data]
  )
  const showTabChrome = visibleTabs.length > 1

  return (
    <SettingsPageProvider
      actionsContainer={actionsContainer}
      titleStatusContainer={titleStatusContainer}
    >
      <div className='flex h-full min-h-0 flex-col gap-4'>
        {showTabChrome ? (
          <div className='flex flex-wrap items-center justify-between gap-2'>
            <span
              ref={setTitleStatusContainer}
              className='inline-flex min-w-0 items-center'
            />
            <div
              ref={setActionsContainer}
              className='flex flex-wrap items-center justify-end gap-2'
            />
          </div>
        ) : null}
        {isLoading ? (
          <div className='text-muted-foreground flex min-h-40 items-center justify-center text-sm'>
            {t('Loading settings...')}
          </div>
        ) : (
          <RatioSettingsCard
            titleKey='Model Pricing'
            modelDefaults={{
              ModelPrice: settings.ModelPrice,
              ModelRatio: settings.ModelRatio,
              CacheRatio: settings.CacheRatio,
              CreateCacheRatio: settings.CreateCacheRatio,
              CompletionRatio: settings.CompletionRatio,
              ImageRatio: settings.ImageRatio,
              AudioRatio: settings.AudioRatio,
              AudioCompletionRatio: settings.AudioCompletionRatio,
              ExposeRatioEnabled: settings.ExposeRatioEnabled,
              BillingMode: settings['billing_setting.billing_mode'],
              BillingExpr: settings['billing_setting.billing_expr'],
            }}
            groupDefaults={{
              TopupGroupRatio: settings.TopupGroupRatio,
              GroupRatio: settings.GroupRatio,
              UserUsableGroups: settings.UserUsableGroups,
              GroupGroupRatio: settings.GroupGroupRatio,
              AutoGroups: settings.AutoGroups,
              DefaultUseAutoGroup: settings.DefaultUseAutoGroup,
              GroupSpecialUsableGroup:
                settings['group_ratio_setting.group_special_usable_group'],
            }}
            toolPricesDefault={settings['tool_price_setting.prices']}
            visibleTabs={visibleTabs}
          />
        )}
      </div>
    </SettingsPageProvider>
  )
}
