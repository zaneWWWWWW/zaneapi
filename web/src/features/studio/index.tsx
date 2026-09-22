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

import { useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { MAX_STUDIO_REFERENCE_IMAGES } from './constants'
import { ImageStudioPanel } from './components/image-studio-panel'
import { StudioGallery } from './components/gallery/studio-gallery'
import { StudioHeader } from './components/studio-header'
import { VideoComingSoonPanel } from './components/video-coming-soon-panel'
import { useStudioState } from './hooks/use-studio-state'
import type { StudioMode } from './types'

interface StudioProps {
  mode: StudioMode
  imageEnabled: boolean
  videoEnabled: boolean
  onModeChange: (mode: StudioMode) => void
}

export function Studio(props: StudioProps) {
  const { t } = useTranslation()
  const state = useStudioState(props.mode)

  const handleReusePrompt = useCallback(
    (prompt: string, model: string, refImage?: string) => {
      state.setPrompt(prompt)
      if (props.mode === 'image') {
        state.setSelectedImageModel(model)
      } else {
        state.setSelectedVideoModel(model)
        if (refImage) {
          state.setReferenceImage(refImage)
        }
        return
      }

      if (!refImage) return
      if (state.referenceImages.length >= MAX_STUDIO_REFERENCE_IMAGES) {
        toast.error(
          t('Maximum of {{count}} reference images', {
            count: MAX_STUDIO_REFERENCE_IMAGES,
          })
        )
        return
      }

      toast.success(t('Added as a reference image'))
      void state.addReferenceFromSrc(refImage).catch(() => {
        toast.error(t('Could not add this image as a reference'))
      })
    },
    [props.mode, state, t]
  )

  return (
    <div className='flex h-[calc(100vh-var(--app-header-height,3.5rem))] w-full flex-col overflow-hidden bg-background'>
      {/* Top Header Bar */}
      <StudioHeader
        mode={props.mode}
        imageEnabled={props.imageEnabled}
        videoEnabled={props.videoEnabled}
        onModeChange={props.onModeChange}
      />

      {/* Main Studio Body: Left Creation Dock + Right Gallery */}
      <div className='flex flex-1 flex-col overflow-hidden lg:flex-row'>
        {/* Left Creation Dock */}
        <div className='w-full shrink-0 border-r border-border/60 bg-card/40 lg:w-[380px] xl:w-[420px]'>
          {props.mode === 'image' ? (
            <ImageStudioPanel
              models={state.currentAvailableModels}
              selectedModel={state.selectedImageModel}
              selectedGroup={state.selectedGroup}
              onModelChange={(group, model) => {
                state.setSelectedGroup(group)
                state.setSelectedImageModel(model)
              }}
              prompt={state.prompt}
              negativePrompt={state.negativePrompt}
              onPromptChange={state.setPrompt}
              onNegativePromptChange={state.setNegativePrompt}
              aspectRatio={state.aspectRatio}
              onAspectRatioChange={state.setAspectRatio}
              imageCount={state.imageCount}
              onImageCountChange={state.setImageCount}
              quality={state.imageQuality}
              onQualityChange={state.setImageQuality}
              style={state.imageStyle}
              onStyleChange={state.setImageStyle}
              referenceImages={state.referenceImages}
              onReferenceImagesChange={state.setReferenceImages}
              isGenerating={state.isGenerating}
              onGenerate={state.handleGenerate}
            />
          ) : (
            <VideoComingSoonPanel />
          )}
        </div>

        {/* Right Gallery Showcase */}
        <div className='flex-1 overflow-hidden'>
          <StudioGallery
            creations={state.creations}
            imageEnabled={props.imageEnabled}
            videoEnabled={props.videoEnabled}
            onDeleteCreation={state.deleteCreation}
            onClearCreations={state.clearCreations}
            onReusePrompt={handleReusePrompt}
          />
        </div>
      </div>
    </div>
  )
}
