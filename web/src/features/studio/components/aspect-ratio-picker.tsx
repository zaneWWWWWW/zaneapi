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

import {
  IMAGE_RATIO_OPTIONS,
  VIDEO_RATIO_OPTIONS,
  type ImageRatioOption,
  type VideoRatioOption,
} from '../constants'
import type { ImageAspectRatio, VideoAspectRatio } from '../types'

interface AspectRatioPickerProps {
  mode: 'image' | 'video'
  imageRatio?: ImageAspectRatio
  videoRatio?: VideoAspectRatio
  onImageRatioChange?: (ratio: ImageAspectRatio) => void
  onVideoRatioChange?: (ratio: VideoAspectRatio) => void
}

export function AspectRatioPicker(props: AspectRatioPickerProps) {
  const { t } = useTranslation()

  if (props.mode === 'image') {
    return (
      <div className='flex flex-col gap-1.5'>
        <label className='text-muted-foreground text-xs font-medium'>
          {t('Aspect Ratio & Resolution')}
        </label>
        <div className='grid grid-cols-4 gap-1.5'>
          {IMAGE_RATIO_OPTIONS.map((option: ImageRatioOption) => {
            const isSelected = props.imageRatio === option.value
            return (
              <button
                key={option.value}
                type='button'
                onClick={() => props.onImageRatioChange?.(option.value)}
                className={`flex flex-col items-center justify-center gap-1 rounded-lg border p-1.5 transition-all ${
                  isSelected
                    ? 'border-primary bg-primary/10 text-primary font-semibold shadow-xs'
                    : 'border-border/60 bg-muted/20 text-muted-foreground hover:border-border hover:bg-muted/40'
                }`}
              >
                <div className='flex h-6 w-6 items-center justify-center'>
                  <div
                    className={`rounded-xs border-2 ${
                      isSelected
                        ? 'border-primary bg-primary/20'
                        : 'border-muted-foreground/50'
                    }`}
                    style={{
                      width: `${option.iconWidth}px`,
                      height: `${option.iconHeight}px`,
                    }}
                  />
                </div>
                <span className='text-xs tabular-nums leading-none'>
                  {option.label}
                </span>
                <span className='text-muted-foreground/70 scale-90 text-[10px] leading-none'>
                  {t(option.descKey)}
                </span>
              </button>
            )
          })}
        </div>
      </div>
    )
  }

  return (
    <div className='flex flex-col gap-1.5'>
      <label className='text-muted-foreground text-xs font-medium'>
        {t('Aspect Ratio')}
      </label>
      <div className='grid grid-cols-3 gap-1.5'>
        {VIDEO_RATIO_OPTIONS.map((option: VideoRatioOption) => {
          const isSelected = props.videoRatio === option.value
          return (
            <button
              key={option.value}
              type='button'
              onClick={() => props.onVideoRatioChange?.(option.value)}
              className={`flex flex-col items-center justify-center gap-1 rounded-lg border p-1.5 transition-all ${
                isSelected
                  ? 'border-primary bg-primary/10 text-primary font-semibold shadow-xs'
                  : 'border-border/60 bg-muted/20 text-muted-foreground hover:border-border hover:bg-muted/40'
              }`}
            >
              <div className='flex h-6 w-6 items-center justify-center'>
                <div
                  className={`rounded-xs border-2 ${
                    isSelected
                      ? 'border-primary bg-primary/20'
                      : 'border-muted-foreground/50'
                  }`}
                  style={{
                    width: `${option.iconWidth}px`,
                    height: `${option.iconHeight}px`,
                  }}
                />
              </div>
              <span className='text-xs tabular-nums leading-none'>
                {option.label}
              </span>
              <span className='text-muted-foreground/70 scale-90 text-[10px] leading-none'>
                {t(option.descKey)}
              </span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
