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
  AlertCircle,
  Copy,
  Download,
  Play,
  RotateCcw,
  Sparkles,
  Trash2,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Progress } from '@/components/ui/progress'
import { Spinner } from '@/components/ui/spinner'

import type { CreationItem } from '../../types'

interface CreationCardProps {
  creation: CreationItem
  onOpenLightbox?: (item: CreationItem) => void
  onOpenVideoModal?: (item: CreationItem) => void
  onReusePrompt?: (prompt: string, model: string, refImage?: string) => void
  onDelete?: (id: string) => void
}

export function CreationCard(props: CreationCardProps) {
  const { t } = useTranslation()
  const item = props.creation

  const handleCopyPrompt = (e: React.MouseEvent) => {
    e.stopPropagation()
    navigator.clipboard.writeText(item.prompt)
    toast.success(t('Prompt copied to clipboard!'))
  }

  const handleDownload = (e: React.MouseEvent) => {
    e.stopPropagation()
    const targetUrl = item.url || (item.b64Json ? `data:image/png;base64,${item.b64Json}` : '')
    if (!targetUrl) return

    const a = document.createElement('a')
    a.href = targetUrl
    a.download = `${item.type}-${item.id}.${item.type === 'video' ? 'mp4' : 'png'}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }

  // Render In-Progress or Queued Card
  if (item.status === 'queued' || item.status === 'in_progress') {
    return (
      <div className='group relative flex flex-col justify-between overflow-hidden rounded-xl border border-primary/30 bg-muted/20 p-4 shadow-xs transition-all hover:border-primary/50'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <Spinner className='size-4 text-primary' />
            <span className='text-xs font-semibold text-primary capitalize'>
              {item.status === 'queued' ? t('Queued in line...') : t('Creating magic...')}
            </span>
          </div>
          <span className='rounded-xs bg-muted/60 px-1.5 py-0.5 text-[10px] text-muted-foreground font-mono'>
            {item.model}
          </span>
        </div>

        <div className='my-6 flex flex-col gap-2'>
          <p className='text-muted-foreground line-clamp-3 text-xs leading-relaxed'>
            {item.prompt}
          </p>
          {item.type === 'video' && typeof item.progress === 'number' && (
            <div className='flex flex-col gap-1'>
              <div className='flex justify-between text-[10px] tabular-nums text-muted-foreground'>
                <span>{t('Rendering progress')}</span>
                <span>{item.progress}%</span>
              </div>
              <Progress value={item.progress} className='h-1.5 w-full' />
            </div>
          )}
        </div>

        <div className='flex items-center justify-between text-[11px] text-muted-foreground/70'>
          <span>{item.type === 'video' ? t('Video Task') : t('Image Generation')}</span>
          <span className='tabular-nums'>
            {new Date(item.createdAt).toLocaleTimeString()}
          </span>
        </div>
      </div>
    )
  }

  // Render Failed Card
  if (item.status === 'failed') {
    return (
      <div className='group relative flex flex-col justify-between overflow-hidden rounded-xl border border-destructive/40 bg-destructive/5 p-4 transition-all'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-1.5 text-xs font-semibold text-destructive'>
            <AlertCircle className='size-4' />
            <span>{t('Generation Failed')}</span>
          </div>
          <button
            type='button'
            onClick={() => props.onDelete?.(item.id)}
            className='text-muted-foreground hover:text-destructive text-xs transition-colors'
            title={t('Delete')}
          >
            <Trash2 className='size-3.5' />
          </button>
        </div>

        <div className='my-4 flex flex-col gap-1.5'>
          <p className='text-foreground/80 line-clamp-2 text-xs font-medium'>
            {item.prompt}
          </p>
          {item.failReason && (
            <p className='text-destructive/80 line-clamp-3 text-[11px] font-mono leading-tight'>
              {t(item.failReason)}
            </p>
          )}
        </div>

        <div className='flex items-center justify-between'>
          <button
            type='button'
            onClick={() => props.onReusePrompt?.(item.prompt, item.model)}
            className='text-primary hover:underline flex items-center gap-1 text-xs'
          >
            <RotateCcw className='size-3' />
            <span>{t('Retry / Reuse')}</span>
          </button>
          <span className='text-[10px] text-muted-foreground tabular-nums'>
            {new Date(item.createdAt).toLocaleTimeString()}
          </span>
        </div>
      </div>
    )
  }

  // Render Completed Card (Image or Video)
  const displayUrl =
    item.url || (item.b64Json ? `data:image/png;base64,${item.b64Json}` : '')

  return (
    <div
      onClick={() => {
        if (item.type === 'video') {
          props.onOpenVideoModal?.(item)
        } else {
          props.onOpenLightbox?.(item)
        }
      }}
      className='group relative flex cursor-pointer flex-col overflow-hidden rounded-xl border border-border/70 bg-card shadow-xs transition-all hover:border-primary/50 hover:shadow-md'
    >
      {/* Media Preview Thumbnail */}
      <div className='relative aspect-square w-full overflow-hidden bg-muted/30'>
        {item.type === 'video' ? (
          <div className='relative flex h-full w-full items-center justify-center bg-black'>
            <video
              src={displayUrl}
              className='h-full w-full object-cover pointer-events-none'
              muted
              playsInline
              preload='metadata'
            />
            <div className='absolute flex h-10 w-10 items-center justify-center rounded-full bg-black/60 text-white shadow-lg backdrop-blur-xs transition-transform group-hover:scale-110'>
              <Play className='size-4 fill-white translate-x-0.5' />
            </div>
            <div className='absolute bottom-2 left-2 rounded-xs bg-black/70 px-1.5 py-0.5 text-[10px] font-medium text-white tabular-nums backdrop-blur-xs'>
              {item.duration || 5}s {item.aspectRatio || '16:9'}
            </div>
          </div>
        ) : (
          <img
            src={displayUrl}
            alt={item.prompt}
            className='h-full w-full object-cover transition-transform duration-300 group-hover:scale-103'
            loading='lazy'
          />
        )}

        {/* Hover Overlay Controls */}
        <div className='absolute inset-0 flex flex-col justify-between bg-gradient-to-t from-black/80 via-black/20 to-transparent p-3 opacity-0 transition-opacity duration-200 group-hover:opacity-100'>
          <div className='flex items-center justify-between'>
            <span className='rounded-xs bg-black/60 px-1.5 py-0.5 text-[10px] font-medium text-white backdrop-blur-xs'>
              {item.model}
            </span>
            <div className='flex items-center gap-1.5'>
              <button
                type='button'
                onClick={handleCopyPrompt}
                className='flex h-7 w-7 items-center justify-center rounded-md bg-black/60 text-white backdrop-blur-xs transition-colors hover:bg-black/90'
                title={t('Copy Prompt')}
              >
                <Copy className='size-3.5' />
              </button>
              <button
                type='button'
                onClick={handleDownload}
                className='flex h-7 w-7 items-center justify-center rounded-md bg-black/60 text-white backdrop-blur-xs transition-colors hover:bg-black/90'
                title={t('Download')}
              >
                <Download className='size-3.5' />
              </button>
              <button
                type='button'
                onClick={(e) => {
                  e.stopPropagation()
                  props.onDelete?.(item.id)
                }}
                className='flex h-7 w-7 items-center justify-center rounded-md bg-black/60 text-white backdrop-blur-xs transition-colors hover:bg-destructive'
                title={t('Delete')}
              >
                <Trash2 className='size-3.5' />
              </button>
            </div>
          </div>

          <div className='flex flex-col gap-1.5'>
            <p className='line-clamp-2 text-xs leading-snug text-white font-medium drop-shadow-xs'>
              {item.prompt}
            </p>
            <div className='flex items-center justify-between'>
              <button
                type='button'
                onClick={(e) => {
                  e.stopPropagation()
                  props.onReusePrompt?.(item.prompt, item.model, displayUrl)
                }}
                className='flex items-center gap-1 text-[11px] font-medium text-primary hover:underline'
              >
                <Sparkles className='size-3' />
                <span>{t('Remix / Reuse')}</span>
              </button>
              <span className='text-[10px] text-white/70'>
                {new Date(item.createdAt).toLocaleTimeString([], {
                  hour: '2-digit',
                  minute: '2-digit',
                })}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
