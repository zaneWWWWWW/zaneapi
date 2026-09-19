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

import { Copy, Download, Sparkles } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

import type { CreationItem } from '../../types'

interface VideoModalProps {
  item: CreationItem | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onReusePrompt?: (prompt: string, model: string) => void
}

export function VideoModal(props: VideoModalProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  if (!props.item) return null
  const item = props.item
  const videoUrl = item.url

  const handleCopy = () => {
    navigator.clipboard.writeText(item.prompt)
    setCopied(true)
    toast.success(t('Prompt copied to clipboard!'))
    setTimeout(() => setCopied(false), 2000)
  }

  const handleDownload = () => {
    if (!videoUrl) return
    const a = document.createElement('a')
    a.href = videoUrl
    a.download = `video-${item.id}.mp4`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className='max-w-4xl overflow-hidden p-0 sm:max-h-[90vh]'>
        <DialogHeader className='hidden'>
          <DialogTitle>{t('Video Player')}</DialogTitle>
        </DialogHeader>

        <div className='flex flex-col md:flex-row md:max-h-[85vh]'>
          {/* Main Video Area */}
          <div className='relative flex flex-1 items-center justify-center bg-black p-4'>
            {videoUrl ? (
              <video
                src={videoUrl}
                controls
                autoPlay
                loop
                playsInline
                className='max-h-[55vh] md:max-h-[80vh] w-auto max-w-full rounded-lg'
              />
            ) : (
              <div className='flex h-64 items-center justify-center text-white/60'>
                {t('Video not available')}
              </div>
            )}
          </div>

          {/* Sidebar Information Area */}
          <div className='flex w-full flex-col justify-between border-t md:border-t-0 md:border-l border-border/70 bg-card p-5 md:w-[320px]'>
            <div className='flex flex-col gap-4'>
              <div className='flex items-center justify-between'>
                <span className='rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary'>
                  {item.model}
                </span>
                <span className='text-[11px] text-muted-foreground tabular-nums'>
                  {new Date(item.createdAt).toLocaleString()}
                </span>
              </div>

              {/* Prompt Section */}
              <div className='flex flex-col gap-1.5'>
                <label className='text-muted-foreground text-xs font-medium'>
                  {t('Prompt')}
                </label>
                <div className='relative rounded-lg border border-border/60 bg-muted/20 p-3'>
                  <p className='text-xs leading-relaxed text-foreground max-h-36 overflow-y-auto'>
                    {item.prompt}
                  </p>
                  <button
                    type='button'
                    onClick={handleCopy}
                    className='text-muted-foreground hover:text-foreground mt-2 flex items-center gap-1 text-[11px] font-medium transition-colors'
                  >
                    <Copy className='size-3' />
                    <span>{copied ? t('Copied!') : t('Copy Prompt')}</span>
                  </button>
                </div>
              </div>

              {/* Metadata */}
              <div className='flex flex-col gap-2 rounded-lg border border-border/60 bg-muted/10 p-3 text-xs'>
                {item.duration && (
                  <div className='flex items-center justify-between text-muted-foreground'>
                    <span>{t('Duration')}</span>
                    <span className='font-mono text-foreground font-medium'>
                      {item.duration}s
                    </span>
                  </div>
                )}
                {item.aspectRatio && (
                  <div className='flex items-center justify-between text-muted-foreground'>
                    <span>{t('Aspect Ratio')}</span>
                    <span className='font-mono text-foreground font-medium'>
                      {item.aspectRatio}
                    </span>
                  </div>
                )}
                {item.taskId && (
                  <div className='flex items-center justify-between text-muted-foreground'>
                    <span>{t('Task ID')}</span>
                    <span className='font-mono text-[11px] text-foreground'>
                      {item.taskId}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Bottom Actions */}
            <div className='mt-6 flex flex-col gap-2'>
              <Button
                type='button'
                variant='outline'
                onClick={() => {
                  props.onOpenChange(false)
                  props.onReusePrompt?.(item.prompt, item.model)
                }}
                className='w-full gap-1.5'
              >
                <Sparkles className='size-3.5 text-primary' />
                <span>{t('Reuse Settings')}</span>
              </Button>
              <Button
                type='button'
                onClick={handleDownload}
                className='w-full gap-1.5'
              >
                <Download className='size-3.5' />
                <span>{t('Download Video')}</span>
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
