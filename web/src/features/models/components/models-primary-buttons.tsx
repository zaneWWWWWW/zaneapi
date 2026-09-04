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
import {
  Plus,
  MoreHorizontal,
  RefreshCw,
  List,
  Building2,
  AlertCircle,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  ADMIN_PERMISSION_ACTIONS,
  ADMIN_PERMISSION_RESOURCES,
  hasPermission,
} from '@/lib/admin-permissions'
import { useAuthStore } from '@/stores/auth-store'

import { useModels } from './models-provider'

export function ModelsPrimaryButtons() {
  const { t } = useTranslation()
  const { setOpen, setCurrentRow } = useModels()
  const currentUser = useAuthStore((s) => s.auth.user)
  const canWrite = hasPermission(
    currentUser,
    ADMIN_PERMISSION_RESOURCES.MODELS,
    ADMIN_PERMISSION_ACTIONS.WRITE
  )

  const handleCreateModel = () => {
    setCurrentRow(null)
    setOpen('create-model')
  }

  const handleMissingModels = () => {
    setOpen('missing-models')
  }

  const handleSync = () => {
    setOpen('sync-wizard')
  }

  const handlePrefillGroups = () => {
    setOpen('prefill-groups')
  }

  const handleManageVendors = () => {
    setOpen('create-vendor') // Will be a separate vendors management dialog
  }

  return (
    <div className='flex items-center gap-2'>
      {canWrite ? (
        <Button onClick={handleCreateModel} size='sm'>
          <Plus className='h-4 w-4' />
          {t('Add Model')}
        </Button>
      ) : null}

      <DropdownMenu>
        <DropdownMenuTrigger render={<Button variant='outline' size='sm' />}>
          <MoreHorizontal className='h-4 w-4' />
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end' className='w-56'>
          <DropdownMenuItem onClick={handleMissingModels}>
            {t('Missing Models')}
            <DropdownMenuShortcut>
              <AlertCircle className='h-4 w-4' />
            </DropdownMenuShortcut>
          </DropdownMenuItem>

          {canWrite ? (
            <DropdownMenuItem onClick={handleSync}>
              {t('Sync Upstream')}
              <DropdownMenuShortcut>
                <RefreshCw className='h-4 w-4' />
              </DropdownMenuShortcut>
            </DropdownMenuItem>
          ) : null}

          {canWrite ? <DropdownMenuSeparator /> : null}

          {canWrite ? (
            <DropdownMenuItem onClick={handlePrefillGroups}>
              {t('Prefill Groups')}
              <DropdownMenuShortcut>
                <List className='h-4 w-4' />
              </DropdownMenuShortcut>
            </DropdownMenuItem>
          ) : null}

          {canWrite ? (
            <DropdownMenuItem onClick={handleManageVendors}>
              {t('Manage Vendors')}
              <DropdownMenuShortcut>
                <Building2 className='h-4 w-4' />
              </DropdownMenuShortcut>
            </DropdownMenuItem>
          ) : null}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
