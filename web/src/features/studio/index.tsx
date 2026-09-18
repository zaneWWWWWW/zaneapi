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

import { ImageStudioPanel } from './components/image-studio-panel'
import { StudioGallery } from './components/gallery/studio-gallery'
import { StudioHeader } from './components/studio-header'
import { VideoNoticeDialog } from './components/video-notice-dialog'
import { VideoStudioPanel } from './components/video-studio-panel'
import { useStudioState } from './hooks/use-studio-state'

export function Studio() {
  const state = useStudioState()

  const handleReusePrompt = useCallback(
    (prompt: string, model: string, refImage?: string) => {
      state.setPrompt(prompt)
      if (refImage) {
        state.setReferenceImage(refImage)
      }
      if (state.mode === 'image') {
        state.setSelectedImageModel(model)
      } else {
        state.setSelectedVideoModel(model)
      }
    },
    [state]
  )

  return (
    <div className='flex h-[calc(100vh-var(--app-header-height,3.5rem))] w-full flex-col overflow-hidden bg-background'>
      {/* Top Header Bar */}
      <StudioHeader
        mode={state.mode}
        onModeChange={state.setMode}
        groups={state.groups}
        selectedGroup={state.selectedGroup}
        onGroupChange={state.setSelectedGroup}
      />

      {/* Main Studio Body: Left Creation Dock + Right Gallery */}
      <div className='flex flex-1 flex-col overflow-hidden lg:flex-row'>
        {/* Left Creation Dock */}
        <div className='w-full shrink-0 border-r border-border/60 bg-card/40 lg:w-[380px] xl:w-[420px]'>
          {state.mode === 'image' ? (
            <ImageStudioPanel
              models={state.currentAvailableModels}
              selectedModel={state.selectedImageModel}
              onModelChange={state.setSelectedImageModel}
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
              referenceImage={state.referenceImage}
              onReferenceImageChange={state.setReferenceImage}
              isGenerating={state.isGenerating}
              onGenerate={state.handleGenerate}
            />
          ) : (
            <VideoStudioPanel
              models={state.currentAvailableModels}
              selectedModel={state.selectedVideoModel}
              onModelChange={state.setSelectedVideoModel}
              prompt={state.prompt}
              onPromptChange={state.setPrompt}
              aspectRatio={state.videoAspectRatio}
              onAspectRatioChange={state.setVideoAspectRatio}
              duration={state.videoDuration}
              onDurationChange={state.setVideoDuration}
              resolution={state.videoResolution}
              onResolutionChange={state.setVideoResolution}
              cameraMotion={state.cameraMotion}
              onCameraMotionChange={state.setCameraMotion}
              referenceImage={state.referenceImage}
              onReferenceImageChange={state.setReferenceImage}
              lastFrameImage={state.lastFrameImage}
              onLastFrameImageChange={state.setLastFrameImage}
              isGenerating={state.isGenerating}
              onGenerate={state.handleGenerate}
            />
          )}
        </div>

        {/* Right Gallery Showcase */}
        <div className='flex-1 overflow-hidden'>
          <StudioGallery
            creations={state.creations}
            onDeleteCreation={state.deleteCreation}
            onClearCreations={state.clearCreations}
            onReusePrompt={handleReusePrompt}
          />
        </div>
      </div>

      {/* Video Generation First-time Notice Dialog */}
      <VideoNoticeDialog
        open={state.videoNoticeOpen}
        onOpenChange={state.setVideoNoticeOpen}
        onConfirm={state.confirmVideoNotice}
      />
    </div>
  )
}
