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

import { Clapperboard } from 'lucide-react'
import { useTranslation } from 'react-i18next'

export function VideoComingSoonPanel() {
  const { t } = useTranslation()

  return (
    <div className='flex h-full flex-col items-center justify-center gap-3 px-8 py-16 text-center'>
      <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary'>
        <Clapperboard className='size-6' aria-hidden='true' />
      </div>
      <span className='text-sm font-semibold'>{t('AI Video')}</span>
      <p className='text-xs font-medium text-muted-foreground'>
        {t('Coming Soon!')}
      </p>
      <p className='text-[11px] text-muted-foreground'>
        {t('Stay tuned though!')}
      </p>
    </div>
  )
}
