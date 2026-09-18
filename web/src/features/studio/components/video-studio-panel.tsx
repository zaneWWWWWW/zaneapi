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
  Clapperboard,
  Clock,
  Compass,
  Film,
  Loader2,
  Tv,
} from 'lucide-react'
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

import {
  CAMERA_MOTIONS,
  VIDEO_DURATIONS,
  VIDEO_RESOLUTIONS,
} from '../constants'
import type {
  CameraMotion,
  ModelOption,
  VideoAspectRatio,
  VideoDuration,
  VideoResolution,
} from '../types'
import { AspectRatioPicker } from './aspect-ratio-picker'
import { PromptEditor } from './prompt-editor'
import { ReferenceImageUpload } from './reference-image-upload'

interface VideoStudioPanelProps {
  models: ModelOption[]
  selectedModel: string
  onModelChange: (model: string) => void
  prompt: string
  onPromptChange: (val: string) => void
  aspectRatio: VideoAspectRatio
  onAspectRatioChange: (val: VideoAspectRatio) => void
  duration: VideoDuration
  onDurationChange: (val: VideoDuration) => void
  resolution: VideoResolution
  onResolutionChange: (val: VideoResolution) => void
  cameraMotion: CameraMotion
  onCameraMotionChange: (val: CameraMotion) => void
  referenceImage: string | null
  onReferenceImageChange: (img: string | null) => void
  lastFrameImage: string | null
  onLastFrameImageChange: (img: string | null) => void
  isGenerating: boolean
  onGenerate: () => void
}

export function VideoStudioPanel(props: VideoStudioPanelProps) {
  const { t } = useTranslation()

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
          <label className='text-muted-foreground text-xs font-medium'>
            {t('Video AI Model')}
          </label>
          <Select
            value={props.selectedModel}
            onValueChange={props.onModelChange}
          >
            <SelectTrigger className='w-full text-xs font-medium'>
              <SelectValue placeholder={t('Select Video Model')} />
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
          mode='video'
          prompt={props.prompt}
          negativePrompt=''
          onPromptChange={props.onPromptChange}
          onNegativePromptChange={() => {}}
        />

        {/* Aspect Ratio */}
        <AspectRatioPicker
          mode='video'
          videoRatio={props.aspectRatio}
          onVideoRatioChange={props.onAspectRatioChange}
        />

        {/* Parameters: Duration & Resolution */}
        <div className='grid grid-cols-2 gap-2.5'>
          {/* Duration */}
          <div className='flex flex-col gap-1.5'>
            <label className='text-muted-foreground flex items-center gap-1 text-xs font-medium'>
              <Clock className='size-3 text-primary' />
              <span>{t('Duration')}</span>
            </label>
            <div className='grid grid-cols-3 gap-1 rounded-lg border border-border/60 p-1 bg-muted/20'>
              {VIDEO_DURATIONS.map((dur) => {
                const isSelected = props.duration === dur
                return (
                  <button
                    key={dur}
                    type='button'
                    onClick={() => props.onDurationChange(dur)}
                    className={`rounded-md py-1 text-xs font-medium transition-all ${
                      isSelected
                        ? 'bg-background text-foreground shadow-xs'
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    {dur}s
                  </button>
                )
              })}
            </div>
          </div>

          {/* Resolution */}
          <div className='flex flex-col gap-1.5'>
            <label className='text-muted-foreground flex items-center gap-1 text-xs font-medium'>
              <Tv className='size-3 text-primary' />
              <span>{t('Resolution')}</span>
            </label>
            <div className='grid grid-cols-2 gap-1 rounded-lg border border-border/60 p-1 bg-muted/20'>
              {VIDEO_RESOLUTIONS.map((res) => {
                const isSelected = props.resolution === res.value
                return (
                  <button
                    key={res.value}
                    type='button'
                    onClick={() => props.onResolutionChange(res.value)}
                    className={`rounded-md py-1 text-xs font-medium transition-all ${
                      isSelected
                        ? 'bg-background text-foreground shadow-xs'
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    {res.label}
                  </button>
                )
              })}
            </div>
          </div>
        </div>

        {/* Camera Motion */}
        <div className='flex flex-col gap-1.5'>
          <label className='text-muted-foreground flex items-center gap-1 text-xs font-medium'>
            <Compass className='size-3 text-primary' />
            <span>{t('Camera Movement')}</span>
          </label>
          <div className='grid grid-cols-5 gap-1'>
            {CAMERA_MOTIONS.map((motion) => {
              const isSelected = props.cameraMotion === motion.value
              return (
                <button
                  key={motion.value}
                  type='button'
                  onClick={() => props.onCameraMotionChange(motion.value)}
                  className={`rounded-lg border py-1.5 text-[11px] font-medium transition-all ${
                    isSelected
                      ? 'border-primary bg-primary/10 text-primary shadow-xs font-semibold'
                      : 'border-border/60 bg-muted/20 text-muted-foreground hover:border-border hover:bg-muted/40'
                  }`}
                >
                  {t(motion.labelKey)}
                </button>
              )
            })}
          </div>
        </div>

        {/* Reference Image Upload (First frame + optional last frame) */}
        <ReferenceImageUpload
          label={t('First-Frame Image (Image-to-Video)')}
          description={t('Upload first frame image to guide video motion')}
          referenceImage={props.referenceImage}
          onImageChange={props.onReferenceImageChange}
          showSecondary
          secondaryLabel={t('Ending Frame Image (Optional)')}
          secondaryImage={props.lastFrameImage}
          onSecondaryImageChange={props.onLastFrameImageChange}
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
              <span>{t('Submitting Task...')}</span>
            </>
          ) : (
            <>
              <Clapperboard className='size-4' />
              <span>{t('Generate Video')}</span>
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
