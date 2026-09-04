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
import type { Row } from '@tanstack/react-table'
import { Pencil, Power, PowerOff, RotateCcw } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  ADMIN_PERMISSION_ACTIONS,
  ADMIN_PERMISSION_RESOURCES,
  hasPermission,
} from '@/lib/admin-permissions'
import { useAuthStore } from '@/stores/auth-store'

import type { PlanRecord } from '../types'
import { useSubscriptions } from './subscriptions-provider'

interface DataTableRowActionsProps {
  row: Row<PlanRecord>
}

export function DataTableRowActions({ row }: DataTableRowActionsProps) {
  const { t } = useTranslation()
  const { setOpen, setCurrentRow, complianceConfirmed } = useSubscriptions()
  const currentUser = useAuthStore((s) => s.auth.user)
  const canWrite = hasPermission(
    currentUser,
    ADMIN_PERMISSION_RESOURCES.SUBSCRIPTIONS,
    ADMIN_PERMISSION_ACTIONS.WRITE
  )
  const canMutate = canWrite && complianceConfirmed
  const isEnabled = row.original.plan.enabled
  const toggleLabel = isEnabled ? t('Disable') : t('Enable')

  const handleEdit = () => {
    if (!canMutate) return
    setCurrentRow(row.original)
    setOpen('update')
  }

  const handleToggleStatus = () => {
    if (!canMutate) return
    setCurrentRow(row.original)
    setOpen('toggle-status')
  }

  const handleResetSubscriptions = () => {
    if (!canMutate) return
    setCurrentRow(row.original)
    setOpen('reset-subscriptions')
  }

  if (!canWrite) {
    return null
  }

  return (
    <div className='-ml-1.5 flex items-center gap-1'>
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant='ghost'
              size='icon-sm'
              disabled={!canMutate}
              onClick={handleEdit}
              aria-label={t('Edit')}
            />
          }
        >
          <Pencil />
        </TooltipTrigger>
        <TooltipContent>{t('Edit')}</TooltipContent>
      </Tooltip>

      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant='ghost'
              size='icon-sm'
              disabled={!canMutate}
              onClick={handleResetSubscriptions}
              aria-label={t('Reset subscription quota')}
            />
          }
        >
          <RotateCcw />
        </TooltipTrigger>
        <TooltipContent>{t('Reset subscription quota')}</TooltipContent>
      </Tooltip>

      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant='ghost'
              size='icon-sm'
              disabled={!canMutate}
              onClick={handleToggleStatus}
              aria-label={toggleLabel}
              className={
                isEnabled
                  ? 'text-destructive hover:text-destructive'
                  : 'text-success hover:text-success'
              }
            />
          }
        >
          {isEnabled ? <PowerOff /> : <Power />}
        </TooltipTrigger>
        <TooltipContent>{toggleLabel}</TooltipContent>
      </Tooltip>
    </div>
  )
}
