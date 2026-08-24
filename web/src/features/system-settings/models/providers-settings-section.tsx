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
import { useTranslation } from 'react-i18next'

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

import type { ModelSettings } from '../types'
import { ClaudeSettingsCard } from './claude-settings-card'
import { GeminiSettingsCard } from './gemini-settings-card'
import { GrokSettingsCard } from './grok-settings-card'

type ProvidersSettingsSectionProps = {
  settings: ModelSettings
}

export function ProvidersSettingsSection(props: ProvidersSettingsSectionProps) {
  const { t } = useTranslation()
  const settings = props.settings

  return (
    <Tabs defaultValue='gemini' className='gap-4'>
      <TabsList className='max-w-full flex-wrap justify-start group-data-horizontal/tabs:h-auto'>
        <TabsTrigger value='gemini'>{t('Gemini')}</TabsTrigger>
        <TabsTrigger value='claude'>{t('Claude')}</TabsTrigger>
        <TabsTrigger value='grok'>{t('Grok')}</TabsTrigger>
      </TabsList>
      <TabsContent value='gemini'>
        <GeminiSettingsCard
          defaultValues={{
            gemini: {
              safety_settings: settings['gemini.safety_settings'],
              version_settings: settings['gemini.version_settings'],
              supported_imagine_models:
                settings['gemini.supported_imagine_models'],
              thinking_adapter_enabled:
                settings['gemini.thinking_adapter_enabled'],
              thinking_adapter_budget_tokens_percentage:
                settings['gemini.thinking_adapter_budget_tokens_percentage'],
              function_call_thought_signature_enabled:
                settings['gemini.function_call_thought_signature_enabled'],
              remove_function_response_id_enabled:
                settings['gemini.remove_function_response_id_enabled'],
            },
          }}
        />
      </TabsContent>
      <TabsContent value='claude'>
        <ClaudeSettingsCard
          defaultValues={{
            claude: {
              model_headers_settings: settings['claude.model_headers_settings'],
              default_max_tokens: settings['claude.default_max_tokens'],
              thinking_adapter_enabled:
                settings['claude.thinking_adapter_enabled'],
              thinking_adapter_budget_tokens_percentage:
                settings['claude.thinking_adapter_budget_tokens_percentage'],
            },
          }}
        />
      </TabsContent>
      <TabsContent value='grok'>
        <GrokSettingsCard
          defaultValues={{
            'grok.violation_deduction_enabled':
              settings['grok.violation_deduction_enabled'] ?? true,
            'grok.violation_deduction_amount':
              settings['grok.violation_deduction_amount'] ?? 0.05,
          }}
        />
      </TabsContent>
    </Tabs>
  )
}
