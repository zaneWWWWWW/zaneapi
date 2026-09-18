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

interface LightboxModalProps {
  item: CreationItem | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onReusePrompt?: (prompt: string, model: string, refImage?: string) => void
}

export function LightboxModal(props: LightboxModalProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  if (!props.item) return null
  const item = props.item
  const displayUrl =
    item.url || (item.b64Json ? `data:image/png;base64,${item.b64Json}` : '')

  const handleCopy = () => {
    navigator.clipboard.writeText(item.prompt)
    setCopied(true)
    toast.success(t('Prompt copied to clipboard!'))
    setTimeout(() => setCopied(false), 2000)
  }

  const handleDownload = () => {
    if (!displayUrl) return
    const a = document.createElement('a')
    a.href = displayUrl
    a.download = `image-${item.id}.png`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className='max-w-4xl overflow-hidden p-0 sm:max-h-[90vh]'>
        <DialogHeader className='hidden'>
          <DialogTitle>{t('Artwork Preview')}</DialogTitle>
        </DialogHeader>

        <div className='flex flex-col md:flex-row md:max-h-[85vh]'>
          {/* Main Image Area */}
          <div className='relative flex flex-1 items-center justify-center bg-black/95 p-4'>
            <img
              src={displayUrl}
              alt={item.prompt}
              className='max-h-[55vh] md:max-h-[80vh] w-auto max-w-full object-contain'
            />
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

              {/* Aspect Ratio and Dimensions */}
              {item.aspectRatio && (
                <div className='flex items-center justify-between text-xs text-muted-foreground'>
                  <span>{t('Aspect Ratio')}</span>
                  <span className='font-mono text-foreground font-medium'>
                    {item.aspectRatio}
                  </span>
                </div>
              )}
            </div>

            {/* Bottom Actions */}
            <div className='mt-6 flex flex-col gap-2'>
              <Button
                type='button'
                variant='outline'
                onClick={() => {
                  props.onOpenChange(false)
                  props.onReusePrompt?.(item.prompt, item.model, displayUrl)
                }}
                className='w-full gap-1.5'
              >
                <Sparkles className='size-3.5 text-primary' />
                <span>{t('Use as Reference / Remix')}</span>
              </Button>
              <Button
                type='button'
                onClick={handleDownload}
                className='w-full gap-1.5'
              >
                <Download className='size-3.5' />
                <span>{t('Download High-Res')}</span>
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
