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
  Clipboard,
  Image as ImageIcon,
  Plus,
  RefreshCw,
  Trash2,
  UploadCloud,
} from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

interface ReferenceImageUploadProps {
  label?: string
  description?: string
  referenceImage: string | null
  onImageChange: (dataUrl: string | null) => void
  secondaryImage?: string | null
  onSecondaryImageChange?: (dataUrl: string | null) => void
  showSecondary?: boolean
  secondaryLabel?: string
}

export function ReferenceImageUpload(props: ReferenceImageUploadProps) {
  const { t } = useTranslation()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const secondaryFileInputRef = useRef<HTMLInputElement>(null)
  const [isDragOver, setIsDragOver] = useState(false)

  const processFile = useCallback(
    (file: File, isSecondary = false) => {
      if (!file.type.startsWith('image/')) {
        toast.error(t('Please select a valid image file'))
        return
      }

      if (file.size > 15 * 1024 * 1024) {
        toast.error(t('Image size exceeds 15MB limit'))
        return
      }

      const reader = new FileReader()
      reader.onload = (e) => {
        if (typeof e.target?.result === 'string') {
          if (isSecondary && props.onSecondaryImageChange) {
            props.onSecondaryImageChange(e.target.result)
          } else {
            props.onImageChange(e.target.result)
          }
        }
      }
      reader.readAsDataURL(file)
    },
    [props, t]
  )

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragOver(false)
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      processFile(e.dataTransfer.files[0])
    }
  }

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragOver(true)
  }

  const handleDragLeave = () => {
    setIsDragOver(false)
  }

  // Support Ctrl+V / Cmd+V paste image directly from clipboard
  useEffect(() => {
    const handlePaste = (e: ClipboardEvent) => {
      const items = e.clipboardData?.items
      if (!items) return

      for (let i = 0; i < items.length; i++) {
        if (items[i].type.startsWith('image/')) {
          const file = items[i].getAsFile()
          if (file) {
            processFile(file)
            toast.success(t('Pasted reference image from clipboard!'))
            break
          }
        }
      }
    }

    window.addEventListener('paste', handlePaste)
    return () => window.removeEventListener('paste', handlePaste)
  }, [processFile, t])

  return (
    <div className='flex flex-col gap-2'>
      <div className='flex items-center justify-between'>
        <label className='text-muted-foreground flex items-center gap-1.5 text-xs font-medium'>
          <ImageIcon className='size-3.5 text-primary' />
          <span>{props.label || t('Reference Image / Initial Frame')}</span>
        </label>
        {props.referenceImage && (
          <button
            type='button'
            onClick={() => props.onImageChange(null)}
            className='text-muted-foreground hover:text-destructive flex items-center gap-1 text-[11px] transition-colors'
          >
            <Trash2 className='size-3' />
            <span>{t('Remove')}</span>
          </button>
        )}
      </div>

      {props.referenceImage ? (
        <div className='group relative flex h-40 w-full items-center justify-center overflow-hidden rounded-xl border border-border/80 bg-muted/20 shadow-xs'>
          <img
            src={props.referenceImage}
            alt='Reference'
            className='h-full w-full object-contain p-2'
          />
          <div className='absolute inset-0 flex items-center justify-center gap-2 bg-black/60 opacity-0 backdrop-blur-xs transition-opacity duration-200 group-hover:opacity-100'>
            <button
              type='button'
              onClick={() => fileInputRef.current?.click()}
              className='flex items-center gap-1 rounded-md bg-white/20 px-2.5 py-1.5 text-xs font-medium text-white transition-colors hover:bg-white/30'
            >
              <RefreshCw className='size-3.5' />
              <span>{t('Replace')}</span>
            </button>
            <button
              type='button'
              onClick={() => props.onImageChange(null)}
              className='flex items-center gap-1 rounded-md bg-destructive/80 px-2.5 py-1.5 text-xs font-medium text-white transition-colors hover:bg-destructive'
            >
              <Trash2 className='size-3.5' />
              <span>{t('Delete')}</span>
            </button>
          </div>
          <input
            ref={fileInputRef}
            type='file'
            accept='image/*'
            className='hidden'
            onChange={(e) => {
              if (e.target.files && e.target.files.length > 0) {
                processFile(e.target.files[0])
              }
            }}
          />
        </div>
      ) : (
        <div
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onClick={() => fileInputRef.current?.click()}
          className={`flex min-h-[135px] cursor-pointer flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed p-4 text-center transition-all ${
            isDragOver
              ? 'border-primary bg-primary/10 shadow-md scale-[1.01]'
              : 'border-border/80 bg-muted/15 hover:border-primary/50 hover:bg-muted/30'
          }`}
        >
          <input
            ref={fileInputRef}
            type='file'
            accept='image/*'
            className='hidden'
            onChange={(e) => {
              if (e.target.files && e.target.files.length > 0) {
                processFile(e.target.files[0])
              }
            }}
          />
          <div className='flex h-11 w-11 items-center justify-center rounded-full bg-primary/10 text-primary shadow-xs'>
            <UploadCloud className='size-6' />
          </div>
          <div className='flex flex-col gap-1 max-w-[260px]'>
            <span className='text-foreground text-xs font-semibold'>
              {props.description || t('Click or drag reference image here')}
            </span>
            <div className='flex items-center justify-center gap-1 text-[11px] text-muted-foreground'>
              <Clipboard className='size-3' />
              <span>{t('Or paste directly via Ctrl+V / ⌘+V')}</span>
            </div>
            <span className='text-muted-foreground/60 text-[10px]'>
              {t('PNG, JPG, WEBP up to 15MB')}
            </span>
          </div>
        </div>
      )}

      {/* Optional Last-Frame Image (For Video Transition) */}
      {props.showSecondary && (
        <div className='mt-2 flex flex-col gap-1.5 border-t border-border/50 pt-2'>
          <div className='flex items-center justify-between'>
            <span className='text-muted-foreground text-[11px] font-medium'>
              {props.secondaryLabel || t('Ending Frame Image (Optional)')}
            </span>
            {props.secondaryImage && (
              <button
                type='button'
                onClick={() => props.onSecondaryImageChange?.(null)}
                className='text-muted-foreground hover:text-destructive text-[10px]'
              >
                {t('Remove')}
              </button>
            )}
          </div>
          {props.secondaryImage ? (
            <div className='relative flex h-24 w-full items-center justify-center overflow-hidden rounded-lg border border-border/70 bg-muted/20'>
              <img
                src={props.secondaryImage}
                alt='End Frame'
                className='h-full w-full object-contain p-1'
              />
            </div>
          ) : (
            <button
              type='button'
              onClick={() => secondaryFileInputRef.current?.click()}
              className='border-border/70 bg-muted/10 text-muted-foreground hover:border-primary/40 hover:bg-muted/20 flex h-14 w-full items-center justify-center gap-1.5 rounded-lg border border-dashed text-xs'
            >
              <Plus className='size-3.5' />
              <span>{t('Add Ending Frame for Smooth Interpolation')}</span>
            </button>
          )}
          <input
            ref={secondaryFileInputRef}
            type='file'
            accept='image/*'
            className='hidden'
            onChange={(e) => {
              if (e.target.files && e.target.files.length > 0) {
                processFile(e.target.files[0], true)
              }
            }}
          />
        </div>
      )}
    </div>
  )
}
