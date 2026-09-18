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
  ChevronDown,
  ChevronRight,
  Lightbulb,
  Sparkles,
  X,
} from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

import { IMAGE_INSPIRATIONS, VIDEO_INSPIRATIONS } from '../constants'
import type { InspirationPrompt } from '../types'

interface PromptEditorProps {
  mode: 'image' | 'video'
  prompt: string
  negativePrompt: string
  onPromptChange: (value: string) => void
  onNegativePromptChange: (value: string) => void
}

export function PromptEditor(props: PromptEditorProps) {
  const { t } = useTranslation()
  const [showNegative, setShowNegative] = useState(false)

  const inspirations =
    props.mode === 'image' ? IMAGE_INSPIRATIONS : VIDEO_INSPIRATIONS
  const [selectedInspirationId, setSelectedInspirationId] = useState<string>('')

  const currentInspiration =
    inspirations.find((item) => item.id === selectedInspirationId) ||
    inspirations[0]

  const handleSelectInspiration = (id: string) => {
    setSelectedInspirationId(id)
    const found = inspirations.find((item) => item.id === id)
    if (found) {
      props.onPromptChange(found.prompt)
    }
  }

  const handleClear = () => {
    props.onPromptChange('')
  }

  return (
    <div className='flex flex-col gap-2'>
      <div className='flex items-center justify-between'>
        <label className='text-foreground flex items-center gap-1.5 text-xs font-medium'>
          <Sparkles className='text-primary size-3.5' />
          {t('Prompt')}
        </label>
        {props.prompt && (
          <button
            type='button'
            onClick={handleClear}
            className='text-muted-foreground hover:text-foreground flex items-center gap-1 text-[11px] transition-colors'
          >
            <X className='size-3' />
            {t('Clear')}
          </button>
        )}
      </div>

      <div className='relative'>
        <Textarea
          value={props.prompt}
          onChange={(e) => props.onPromptChange(e.target.value)}
          placeholder={
            props.mode === 'image'
              ? t('Describe the image you want to generate in detail...')
              : t('Describe the video scene, motion, and atmosphere in detail...')
          }
          className='border-border/70 focus-visible:ring-primary min-h-[90px] resize-y rounded-lg text-xs leading-relaxed'
          rows={3}
        />
        <div className='text-muted-foreground/60 pointer-events-none absolute right-2.5 bottom-2 text-[10px] tabular-nums'>
          {props.prompt.length}
        </div>
      </div>

      {/* Borderless Inspiration Dropdown (Automatically fills prompt upon selection) */}
      <div className='flex items-center'>
        <Select
          value={currentInspiration?.id || ''}
          onValueChange={handleSelectInspiration}
        >
          <SelectTrigger className='h-7.5 w-full border-none bg-muted/25 hover:bg-muted/40 text-xs px-2.5 shadow-none rounded-lg focus-visible:ring-0 focus:ring-0'>
            <div className='flex items-center gap-1.5 truncate text-muted-foreground'>
              <Lightbulb className='size-3.5 text-amber-500 shrink-0' />
              <span className='text-muted-foreground/70 text-[11px] shrink-0'>
                {t('Inspiration')}:
              </span>
              <span className='truncate text-foreground font-medium text-xs'>
                {currentInspiration
                  ? `[${t(currentInspiration.categoryKey)}] ${t(currentInspiration.titleKey)}`
                  : t('Select inspiration prompt...')}
              </span>
            </div>
          </SelectTrigger>
          <SelectContent className='max-h-72 w-80'>
            {inspirations.map((item: InspirationPrompt) => (
              <SelectItem
                key={item.id}
                value={item.id}
                className='text-xs cursor-pointer py-1.5'
              >
                <div className='flex items-center gap-1.5'>
                  <span className='text-muted-foreground/70 font-mono text-[10px]'>
                    [{t(item.categoryKey)}]
                  </span>
                  <span className='font-medium'>{t(item.titleKey)}</span>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Negative Prompt (Collapsible) */}
      <div className='mt-0.5 flex flex-col gap-1.5'>
        <button
          type='button'
          onClick={() => setShowNegative(!showNegative)}
          className='text-muted-foreground hover:text-foreground flex w-fit items-center gap-1 text-xs font-medium transition-colors'
        >
          {showNegative ? (
            <ChevronDown className='size-3.5' />
          ) : (
            <ChevronRight className='size-3.5' />
          )}
          <span>{t('Negative Prompt')}</span>
          <span className='text-muted-foreground/60 text-[10px]'>
            ({t('Optional')})
          </span>
        </button>

        {showNegative && (
          <Textarea
            value={props.negativePrompt}
            onChange={(e) => props.onNegativePromptChange(e.target.value)}
            placeholder={t('Elements, flaws, or styles you want to avoid...')}
            className='border-border/60 min-h-[55px] resize-y rounded-lg text-xs leading-relaxed'
            rows={2}
          />
        )}
      </div>
    </div>
  )
}
