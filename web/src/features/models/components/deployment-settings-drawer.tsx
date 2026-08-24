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

import {
  sideDrawerContentClassName,
  sideDrawerFormClassName,
  sideDrawerHeaderClassName,
} from '@/components/drawer-layout'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { SettingsPageProvider } from '@/features/system-settings/components/settings-page-context'
import {
  getOptionValue,
  useSystemOptions,
} from '@/features/system-settings/hooks/use-system-options'
import { IoNetDeploymentSettingsSection } from '@/features/system-settings/integrations/ionet-deployment-settings-section'

const DEPLOYMENT_OPTION_DEFAULTS = {
  'model_deployment.ionet.enabled': false,
  'model_deployment.ionet.api_key': '',
}

type DeploymentSettingsDrawerProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function DeploymentSettingsDrawer(props: DeploymentSettingsDrawerProps) {
  const { t } = useTranslation()
  const { data, isLoading } = useSystemOptions()
  const [actionsContainer, setActionsContainer] =
    useState<HTMLDivElement | null>(null)

  const settings = useMemo(
    () => getOptionValue(data?.data, DEPLOYMENT_OPTION_DEFAULTS),
    [data?.data]
  )

  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent className={sideDrawerContentClassName('sm:max-w-md')}>
        <SheetHeader className={sideDrawerHeaderClassName()}>
          <SheetTitle>{t('Deployment settings')}</SheetTitle>
        </SheetHeader>
        <SettingsPageProvider actionsContainer={actionsContainer}>
          <div className={sideDrawerFormClassName()}>
            {isLoading ? (
              <div className='text-muted-foreground flex min-h-40 items-center justify-center text-sm'>
                {t('Loading settings...')}
              </div>
            ) : (
              <IoNetDeploymentSettingsSection
                defaultValues={{
                  enabled: settings['model_deployment.ionet.enabled'],
                  apiKey: settings['model_deployment.ionet.api_key'],
                }}
              />
            )}
          </div>
          <div
            ref={setActionsContainer}
            className='border-border/70 flex shrink-0 flex-wrap items-center justify-end gap-2 border-t px-4 py-3 sm:px-6'
          />
        </SettingsPageProvider>
      </SheetContent>
    </Sheet>
  )
}
