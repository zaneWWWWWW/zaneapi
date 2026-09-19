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

import { Info, Loader2, Sparkles, Wand2 } from 'lucide-react'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

import { isDalle3 } from '../constants'
import type { ImageAspectRatio, ImageQuality, ImageStyle, ModelOption } from '../types'
import { AspectRatioPicker } from './aspect-ratio-picker'
import { PromptEditor } from './prompt-editor'
import { ReferenceImageUpload } from './reference-image-upload'

interface ImageStudioPanelProps {
  models: ModelOption[]
  selectedModel: string
  onModelChange: (model: string) => void
  prompt: string
  negativePrompt: string
  onPromptChange: (val: string) => void
  onNegativePromptChange: (val: string) => void
  aspectRatio: ImageAspectRatio
  onAspectRatioChange: (val: ImageAspectRatio) => void
  imageCount: number
  onImageCountChange: (count: number) => void
  quality: ImageQuality
  onQualityChange: (q: ImageQuality) => void
  style: ImageStyle
  onStyleChange: (s: ImageStyle) => void
  referenceImage: string | null
  onReferenceImageChange: (img: string | null) => void
  isGenerating: boolean
  onGenerate: () => void
}

export function ImageStudioPanel(props: ImageStudioPanelProps) {
  const { t } = useTranslation()
  const isDalle = isDalle3(props.selectedModel)

  // Auto-adjust count if model is DALL-E 3
  useEffect(() => {
    if (isDalle && props.imageCount > 1) {
      props.onImageCountChange(1)
    }
  }, [isDalle, props])

  // Keyboard shortcut: Cmd/Ctrl + Enter to trigger generate
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        if (!props.isGenerating && props.prompt.trim()) {
          props.onGenerate()
        }
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [props])

  return (
    <div className='flex h-full flex-col justify-between overflow-y-auto p-4'>
      <div className='flex flex-col gap-3.5'>
        {/* Model Selection */}
        <div className='flex flex-col gap-1.5'>
          <div className='flex items-center justify-between'>
            <label className='text-muted-foreground text-xs font-medium'>
              {t('AI Model')}
            </label>
            {isDalle && (
              <span className='text-muted-foreground/80 flex items-center gap-1 text-[10px]'>
                <Info className='size-3 text-amber-500' />
                <span>{t('DALL-E 3 supports 1 image per request')}</span>
              </span>
            )}
          </div>
          <Select
            value={props.selectedModel}
            onValueChange={(v) => v !== null && props.onModelChange(v)}
          >
            <SelectTrigger className='w-full text-xs font-medium'>
              <SelectValue placeholder={t('Select Model')} />
            </SelectTrigger>
            <SelectContent>
              {props.models.map((m) => (
                <SelectItem key={m.value} value={m.value} className='text-xs'>
                  <div className='flex items-center gap-2'>
                    <span>{m.label}</span>
                    {m.isPopular && (
                      <span className='rounded-xs bg-primary/10 px-1 py-0.2 text-[10px] text-primary font-medium'>
                        {t('Popular')}
                      </span>
                    )}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Prompt Editor */}
        <PromptEditor
          mode='image'
          prompt={props.prompt}
          negativePrompt={props.negativePrompt}
          onPromptChange={props.onPromptChange}
          onNegativePromptChange={props.onNegativePromptChange}
        />

        {/* Aspect Ratio & Resolution */}
        <AspectRatioPicker
          mode='image'
          imageRatio={props.aspectRatio}
          onImageRatioChange={props.onAspectRatioChange}
        />

        {/* Parameters: Count & Quality & Style */}
        <div className='grid grid-cols-2 gap-2.5'>
          {/* Batch Count */}
          <div className='flex flex-col gap-1.5'>
            <label className='text-muted-foreground text-xs font-medium'>
              {t('Batch Count')}
            </label>
            <div className='grid grid-cols-4 gap-1 rounded-lg border border-border/60 p-1 bg-muted/20'>
              {[1, 2, 3, 4].map((count) => {
                const disabled = isDalle && count > 1
                let stateClass = 'text-muted-foreground hover:text-foreground'
                if (props.imageCount === count) {
                  stateClass = 'bg-background text-foreground shadow-xs'
                } else if (disabled) {
                  stateClass = 'opacity-30 cursor-not-allowed text-muted-foreground'
                }
                return (
                  <button
                    key={count}
                    type='button'
                    disabled={disabled}
                    onClick={() => props.onImageCountChange(count)}
                    className={`rounded-md py-1 text-xs font-medium transition-all ${stateClass}`}
                  >
                    {count}
                  </button>
                )
              })}
            </div>
          </div>

          {/* Quality */}
          <div className='flex flex-col gap-1.5'>
            <label className='text-muted-foreground text-xs font-medium'>
              {t('Quality')}
            </label>
            <div className='grid grid-cols-3 gap-1 rounded-lg border border-border/60 p-1 bg-muted/20'>
              {(
                [
                  { value: 'standard', labelKey: 'Standard' },
                  { value: 'hd', labelKey: 'HD' },
                  { value: 'ultra', labelKey: 'Ultra' },
                ] as const
              ).map((q) => (
                <button
                  key={q.value}
                  type='button'
                  onClick={() => props.onQualityChange(q.value)}
                  className={`rounded-md py-1 text-xs font-medium transition-all ${
                    props.quality === q.value
                      ? 'bg-background text-foreground shadow-xs'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  {t(q.labelKey)}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Style Selection */}
        <div className='flex flex-col gap-1.5'>
          <label className='text-muted-foreground flex items-center gap-1 text-xs font-medium'>
            <Sparkles className='size-3 text-primary' />
            <span>{t('Rendering Style')}</span>
          </label>
          <div className='grid grid-cols-3 gap-1.5'>
            {(
              [
                { value: 'vivid', labelKey: 'Vivid' },
                { value: 'natural', labelKey: 'Natural' },
                { value: 'anime', labelKey: 'Anime' },
              ] as const
            ).map((s) => (
              <button
                key={s.value}
                type='button'
                onClick={() => props.onStyleChange(s.value)}
                className={`rounded-lg border py-1.5 text-xs font-medium transition-all ${
                  props.style === s.value
                    ? 'border-primary bg-primary/10 text-primary shadow-xs'
                    : 'border-border/60 bg-muted/20 text-muted-foreground hover:border-border hover:bg-muted/40'
                }`}
              >
                {t(s.labelKey)}
              </button>
            ))}
          </div>
        </div>

        {/* Reference Image Uploader (Expanded height) */}
        <ReferenceImageUpload
          label={t('Reference Image / Image-to-Image (Optional)')}
          description={t('Upload reference image to guide style or structure')}
          referenceImage={props.referenceImage}
          onImageChange={props.onReferenceImageChange}
        />
      </div>

      {/* Bottom Generate Button */}
      <div className='sticky bottom-0 mt-6 border-t border-border/60 bg-background/95 pt-3 backdrop-blur-xs'>
        <Button
          type='button'
          onClick={props.onGenerate}
          disabled={props.isGenerating || !props.prompt.trim()}
          className='w-full gap-2 rounded-lg py-5 text-sm font-semibold shadow-md'
        >
          {props.isGenerating ? (
            <>
              <Loader2 className='size-4 animate-spin' />
              <span>{t('Generating Artwork...')}</span>
            </>
          ) : (
            <>
              <Wand2 className='size-4' />
              <span>{t('Generate Image')}</span>
              <span className='text-primary-foreground/70 ml-1 text-xs font-normal'>
                (⌘/Ctrl + ↵)
              </span>
            </>
          )}
        </Button>
      </div>
    </div>
  )
}
