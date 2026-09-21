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

import { Clapperboard, Image as ImageIcon, Layers } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

import type { GroupOption, StudioMode } from '../types'

interface StudioHeaderProps {
  mode: StudioMode
  imageEnabled?: boolean
  videoEnabled?: boolean
  onModeChange: (mode: StudioMode) => void
  groups: GroupOption[]
  selectedGroup: string
  onGroupChange: (group: string) => void
}

export function StudioHeader(props: StudioHeaderProps) {
  const { t } = useTranslation()
  const imageEnabled = props.imageEnabled !== false
  const videoEnabled = props.videoEnabled !== false
  const showModeSwitcher = imageEnabled && videoEnabled

  return (
    <div className='flex h-14 shrink-0 items-center justify-between border-b border-border/60 bg-background/80 px-4 backdrop-blur-md'>
      {showModeSwitcher ? (
        <div className='flex items-center rounded-lg border border-border/70 bg-muted/30 p-1'>
          <button
            type='button'
            onClick={() => props.onModeChange('image')}
            className={`flex items-center gap-2 rounded-md px-3.5 py-1.5 text-xs font-semibold transition-all ${
              props.mode === 'image'
                ? 'bg-background text-foreground shadow-xs'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            <ImageIcon className='size-3.5 text-primary' />
            <span>{t('AI Image')}</span>
          </button>
          <button
            type='button'
            onClick={() => props.onModeChange('video')}
            className={`flex items-center gap-2 rounded-md px-3.5 py-1.5 text-xs font-semibold transition-all ${
              props.mode === 'video'
                ? 'bg-background text-foreground shadow-xs'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            <Clapperboard className='size-3.5 text-primary' />
            <span>{t('AI Video')}</span>
          </button>
        </div>
      ) : (
        <div />
      )}

      {/* Group Selector */}
      {props.groups.length > 0 && (
        <div className='flex items-center gap-2'>
          <div className='flex items-center gap-1.5 text-xs text-muted-foreground'>
            <Layers className='size-3.5' />
            <span className='hidden sm:inline'>{t('Group')}:</span>
          </div>
          <Select
            value={props.selectedGroup}
            onValueChange={(v) => v !== null && props.onGroupChange(v)}
          >
            <SelectTrigger className='h-8 min-w-[110px] text-xs'>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {props.groups.map((g) => (
                <SelectItem key={g.value} value={g.value} className='text-xs'>
                  {g.label} ({g.ratio}x)
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}
    </div>
  )
}
