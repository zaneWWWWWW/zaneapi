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
import { useQueryClient } from '@tanstack/react-query'
import { getRouteApi } from '@tanstack/react-router'
import { Plus, Settings } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

import { listDeployments } from './api'
import { DeploymentAccessGuard } from './components/deployment-access-guard'
import { DeploymentSettingsDrawer } from './components/deployment-settings-drawer'
import { DeploymentsTable } from './components/deployments-table'
import { CreateDeploymentDrawer } from './components/dialogs/create-deployment-drawer'
import { ModelPricingSection } from './components/model-pricing-section'
import { ModelsDialogs } from './components/models-dialogs'
import { ModelsPrimaryButtons } from './components/models-primary-buttons'
import { ModelsProvider, useModels } from './components/models-provider'
import { ModelsTable } from './components/models-table'
import {
  MODEL_WORKSPACE_DEFAULT_VIEW,
  MODEL_WORKSPACE_VIEWS,
  type ModelWorkspaceViewId,
} from './constants'
import { useModelDeploymentSettings } from './hooks/use-model-deployment-settings'
import { deploymentsQueryKeys } from './lib'
import {
  type ModelsSectionId,
  MODELS_DEFAULT_SECTION,
  MODELS_SECTION_IDS,
} from './section-registry'

const route = getRouteApi('/_authenticated/models/$section')

const SECTION_META: Record<ModelsSectionId, { titleKey: string }> = {
  metadata: {
    titleKey: 'Models',
  },
  deployments: {
    titleKey: 'Deployments',
  },
}

function ModelsContent() {
  const { t } = useTranslation()
  const navigate = route.useNavigate()
  const { tabCategory, setTabCategory } = useModels()
  const params = route.useParams()
  const search = route.useSearch()
  const activeSection = (params.section ??
    MODELS_DEFAULT_SECTION) as ModelsSectionId
  const workspaceView = search.view ?? MODEL_WORKSPACE_DEFAULT_VIEW

  const [createDeploymentOpen, setCreateDeploymentOpen] = useState(false)
  const [deploymentSettingsOpen, setDeploymentSettingsOpen] = useState(false)
  const deployment = useModelDeploymentSettings()

  // keep context state in sync (for components that rely on it)
  useEffect(() => {
    if (tabCategory !== activeSection) {
      setTabCategory(activeSection)
    }
  }, [activeSection, setTabCategory, tabCategory])

  const handleSectionChange = useCallback(
    (section: string) => {
      void navigate({
        to: '/models/$section',
        params: { section: section as ModelsSectionId },
      })
    },
    [navigate]
  )

  const handleWorkspaceViewChange = useCallback(
    (view: string) => {
      void navigate({
        to: '/models/$section',
        params: { section: 'metadata' },
        search: (prev) => ({
          ...prev,
          view:
            view === MODEL_WORKSPACE_DEFAULT_VIEW
              ? undefined
              : (view as ModelWorkspaceViewId),
        }),
      })
    },
    [navigate]
  )

  const meta = SECTION_META[activeSection] ?? SECTION_META.metadata
  const activeWorkspaceView =
    MODEL_WORKSPACE_VIEWS.find((view) => view.id === workspaceView) ??
    MODEL_WORKSPACE_VIEWS[0]
  const pricingTab =
    'pricingTab' in activeWorkspaceView
      ? activeWorkspaceView.pricingTab
      : undefined
  let sectionActions = null
  if (activeSection === 'metadata' && workspaceView === 'catalog') {
    sectionActions = <ModelsPrimaryButtons />
  } else if (activeSection === 'deployments') {
    sectionActions = (
      <>
        <Button
          variant='outline'
          size='sm'
          onClick={() => setDeploymentSettingsOpen(true)}
        >
          <Settings className='h-4 w-4' />
          {t('Deployment settings')}
        </Button>
        <Button onClick={() => setCreateDeploymentOpen(true)} size='sm'>
          <Plus className='h-4 w-4' />
          {t('Create deployment')}
        </Button>
      </>
    )
  }
  let sectionContent = (
    <DeploymentsSection
      deployment={deployment}
      onOpenSettings={() => setDeploymentSettingsOpen(true)}
    />
  )
  if (activeSection === 'metadata') {
    sectionContent =
      pricingTab == null ? (
        <ModelsTable />
      ) : (
        <ModelPricingSection visibleTabs={[pricingTab]} />
      )
  }

  return (
    <>
      <SectionPageLayout fixedContent>
        <SectionPageLayout.Title>{t(meta.titleKey)}</SectionPageLayout.Title>
        {sectionActions != null && (
          <SectionPageLayout.Actions>
            {sectionActions}
          </SectionPageLayout.Actions>
        )}
        <SectionPageLayout.Content>
          <div className='flex h-full min-h-0 flex-col gap-4'>
            <div className='flex flex-col gap-2'>
              <Tabs value={activeSection} onValueChange={handleSectionChange}>
                <TabsList className='max-w-full flex-wrap justify-start group-data-horizontal/tabs:h-auto'>
                  {MODELS_SECTION_IDS.map((section) => (
                    <TabsTrigger key={section} value={section}>
                      {t(SECTION_META[section].titleKey)}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </Tabs>
              {activeSection === 'metadata' ? (
                <Tabs
                  value={workspaceView}
                  onValueChange={handleWorkspaceViewChange}
                >
                  <TabsList className='max-w-full flex-wrap justify-start group-data-horizontal/tabs:h-auto'>
                    {MODEL_WORKSPACE_VIEWS.map((view) => (
                      <TabsTrigger key={view.id} value={view.id}>
                        {t(view.titleKey)}
                      </TabsTrigger>
                    ))}
                  </TabsList>
                </Tabs>
              ) : null}
            </div>
            <div className='min-h-0 flex-1'>{sectionContent}</div>
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <ModelsDialogs />
      <CreateDeploymentDrawer
        open={createDeploymentOpen}
        onOpenChange={setCreateDeploymentOpen}
      />
      <DeploymentSettingsDrawer
        open={deploymentSettingsOpen}
        onOpenChange={(open) => {
          setDeploymentSettingsOpen(open)
          if (!open) void deployment.refresh()
        }}
      />
    </>
  )
}

function DeploymentsSection({
  deployment,
  onOpenSettings,
}: {
  deployment: ReturnType<typeof useModelDeploymentSettings>
  onOpenSettings: () => void
}) {
  const queryClient = useQueryClient()
  const {
    loading: deploymentLoading,
    loadingPhase,
    isIoNetEnabled,
    connectionLoading,
    connectionOk,
    connectionError,
    testConnection,
  } = deployment

  // Prefetch deployments list while connection check is in progress.
  useEffect(() => {
    if (isIoNetEnabled && loadingPhase === 'connection') {
      const defaultParams = { p: 1, page_size: 10 }
      queryClient.prefetchQuery({
        queryKey: deploymentsQueryKeys.list(defaultParams),
        queryFn: () => listDeployments(defaultParams),
        staleTime: 30 * 1000,
      })
    }
  }, [isIoNetEnabled, loadingPhase, queryClient])

  return (
    <DeploymentAccessGuard
      loading={deploymentLoading}
      loadingPhase={loadingPhase}
      isEnabled={isIoNetEnabled}
      connectionLoading={connectionLoading}
      connectionOk={connectionOk}
      connectionError={connectionError}
      onRetry={testConnection}
      onOpenSettings={onOpenSettings}
    >
      <DeploymentsTable />
    </DeploymentAccessGuard>
  )
}

export function Models() {
  return (
    <ModelsProvider>
      <ModelsContent />
    </ModelsProvider>
  )
}
