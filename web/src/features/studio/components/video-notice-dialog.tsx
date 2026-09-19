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

import { AlertCircle, Clapperboard } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface VideoNoticeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}

export function VideoNoticeDialog(props: VideoNoticeDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className='max-w-md'>
        <DialogHeader className='flex flex-col gap-2'>
          <div className='flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary'>
            <Clapperboard className='size-6' />
          </div>
          <DialogTitle className='text-base font-semibold'>
            {t('AI Video Generation Notice')}
          </DialogTitle>
          <DialogDescription className='text-xs leading-relaxed text-muted-foreground'>
            {t(
              'AI video rendering typically takes 1 to 3 minutes depending on upstream queue traffic. Please do not close or refresh this page while processing.'
            )}
          </DialogDescription>
        </DialogHeader>

        <div className='flex items-start gap-2.5 rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-xs text-amber-600 dark:text-amber-400'>
          <AlertCircle className='size-4 shrink-0 mt-0.5' />
          <span>
            {t(
              'A browser confirmation prompt will appear if you attempt to leave or reload the tab while generation is in progress.'
            )}
          </span>
        </div>

        <DialogFooter className='mt-2'>
          <Button
            type='button'
            variant='outline'
            size='sm'
            onClick={() => props.onOpenChange(false)}
          >
            {t('Cancel')}
          </Button>
          <Button
            type='button'
            size='sm'
            onClick={() => {
              props.onOpenChange(false)
              props.onConfirm()
            }}
          >
            {t('Understood, Generate Now')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
