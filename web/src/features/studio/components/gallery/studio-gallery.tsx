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
  Image as ImageIcon,
  Info,
  LayoutGrid,
  Sparkles,
  Trash2,
} from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

import type { CreationItem } from '../../types'
import { CreationCard } from './creation-card'
import { LightboxModal } from './lightbox-modal'
import { VideoModal } from './video-modal'

interface StudioGalleryProps {
  creations: CreationItem[]
  onDeleteCreation?: (id: string) => void
  onClearCreations?: () => void
  onReusePrompt?: (prompt: string, model: string, refImage?: string) => void
}

type FilterType = 'all' | 'image' | 'video'

export function StudioGallery(props: StudioGalleryProps) {
  const { t } = useTranslation()
  const [filter, setFilter] = useState<FilterType>('all')
  const [activeLightboxItem, setActiveLightboxItem] =
    useState<CreationItem | null>(null)
  const [activeVideoItem, setActiveVideoItem] = useState<CreationItem | null>(
    null
  )

  const filteredItems = useMemo(() => {
    if (filter === 'all') return props.creations
    return props.creations.filter((item) => item.type === filter)
  }, [filter, props.creations])

  return (
    <div className='flex h-full flex-col overflow-hidden bg-muted/10'>
      {/* Gallery Header Bar */}
      <div className='flex h-14 shrink-0 items-center justify-between border-b border-border/60 bg-background/50 px-4 backdrop-blur-xs'>
        <div className='flex items-center gap-2'>
          <LayoutGrid className='text-muted-foreground size-4' />
          <span className='text-foreground text-xs font-semibold'>
            {t('Creations')}
          </span>
          <span className='rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium text-muted-foreground tabular-nums'>
            {filteredItems.length}
          </span>
        </div>

        <div className='flex items-center gap-2'>
          {/* Filter Pills */}
          <div className='flex items-center rounded-lg border border-border/60 bg-muted/30 p-0.5 text-xs'>
            <button
              type='button'
              onClick={() => setFilter('all')}
              className={`rounded-md px-2.5 py-1 text-xs font-medium transition-all ${
                filter === 'all'
                  ? 'bg-background text-foreground shadow-xs'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {t('All')}
            </button>
            <button
              type='button'
              onClick={() => setFilter('image')}
              className={`flex items-center gap-1 rounded-md px-2.5 py-1 text-xs font-medium transition-all ${
                filter === 'image'
                  ? 'bg-background text-foreground shadow-xs'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              <ImageIcon className='size-3' />
              <span>{t('Images')}</span>
            </button>
            <button
              type='button'
              onClick={() => setFilter('video')}
              className={`flex items-center gap-1 rounded-md px-2.5 py-1 text-xs font-medium transition-all ${
                filter === 'video'
                  ? 'bg-background text-foreground shadow-xs'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              <Clapperboard className='size-3' />
              <span>{t('Videos')}</span>
            </button>
          </div>

          {/* Clear Button */}
          {props.creations.length > 0 && (
            <Button
              type='button'
              variant='ghost'
              size='sm'
              onClick={props.onClearCreations}
              className='text-muted-foreground hover:text-destructive h-7 gap-1 px-2 text-xs'
              title={t('Clear History')}
            >
              <Trash2 className='size-3.5' />
              <span className='hidden sm:inline'>{t('Clear')}</span>
            </Button>
          )}
        </div>
      </div>

      <div className='flex shrink-0 items-start gap-2 border-b border-border/60 bg-muted/20 px-4 py-2'>
        <Info className='text-muted-foreground mt-0.5 size-3.5 shrink-0' />
        <p className='text-muted-foreground text-[11px] leading-relaxed'>
          {t(
            'Generated images are stored only in this browser. Download them promptly after generation.'
          )}
        </p>
      </div>

      {/* Gallery Content Area */}
      <div className='flex-1 overflow-y-auto p-4'>
        {filteredItems.length === 0 ? (
          /* Empty State */
          <div className='flex h-full min-h-[400px] flex-col items-center justify-center gap-3 text-center p-6'>
            <div className='flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary shadow-xs'>
              <Sparkles className='size-7' />
            </div>
            <div className='flex flex-col gap-1 max-w-sm'>
              <h3 className='text-foreground text-sm font-semibold'>
                {t('No creations yet')}
              </h3>
              <p className='text-muted-foreground text-xs leading-relaxed'>
                {t(
                  'Enter your prompt on the left to start generating high-quality AI images or cinematic videos.'
                )}
              </p>
            </div>
          </div>
        ) : (
          /* Responsive Cards Grid */
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'>
            {filteredItems.map((item) => (
              <CreationCard
                key={item.id}
                creation={item}
                onOpenLightbox={(c) => setActiveLightboxItem(c)}
                onOpenVideoModal={(c) => setActiveVideoItem(c)}
                onReusePrompt={props.onReusePrompt}
                onDelete={props.onDeleteCreation}
              />
            ))}
          </div>
        )}
      </div>

      {/* Modals */}
      <LightboxModal
        item={activeLightboxItem}
        open={Boolean(activeLightboxItem)}
        onOpenChange={(open) => !open && setActiveLightboxItem(null)}
        onReusePrompt={props.onReusePrompt}
      />

      <VideoModal
        item={activeVideoItem}
        open={Boolean(activeVideoItem)}
        onOpenChange={(open) => !open && setActiveVideoItem(null)}
        onReusePrompt={props.onReusePrompt}
      />
    </div>
  )
}
