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
import { toIntlLocale } from '@/i18n/languages'

export function getQuotaDataCutoff(now = Date.now()): Date {
  const cutoff = new Date(now)
  cutoff.setMinutes(0, 0, 0)
  return cutoff
}

export function formatQuotaDataCutoff(
  cutoff: Date,
  language?: string
): string {
  return new Intl.DateTimeFormat(toIntlLocale(language), {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(cutoff)
}

export function readDataExportIntervalMinutes(status: unknown): number | null {
  if (!status || typeof status !== 'object') return null
  const record = status as Record<string, unknown>
  const nested =
    record.data && typeof record.data === 'object'
      ? (record.data as Record<string, unknown>)
      : undefined
  const raw = record.data_export_interval ?? nested?.data_export_interval
  const minutes = typeof raw === 'number' ? raw : Number(raw)
  if (!Number.isFinite(minutes) || minutes < 1) return null
  return Math.floor(minutes)
}

export function formatQuotaDataFreshnessMessage(
  t: (key: string, options?: Record<string, unknown>) => string,
  options: { language?: string; status?: unknown; now?: number }
): string {
  const time = formatQuotaDataCutoff(
    getQuotaDataCutoff(options.now),
    options.language
  )
  const minutes = readDataExportIntervalMinutes(options.status)
  if (minutes) {
    return t(
      'Data through {{time}}. Updated about every {{minutes}} minutes.',
      { time, minutes }
    )
  }
  return t('Data through {{time}}.', { time })
}
