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

function readStatusValue(status: unknown, key: string): unknown {
  if (!status || typeof status !== 'object') return undefined
  const record = status as Record<string, unknown>
  if (record[key] !== undefined) return record[key]
  if (!record.data || typeof record.data !== 'object') return undefined
  return (record.data as Record<string, unknown>)[key]
}

export function formatQuotaDataExportTime(
  exportedAt: number,
  language?: string
): string {
  return new Intl.DateTimeFormat(toIntlLocale(language), {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(exportedAt)
}

export function readDataExportIntervalMinutes(status: unknown): number | null {
  const raw = readStatusValue(status, 'data_export_interval')
  const minutes = typeof raw === 'number' ? raw : Number(raw)
  if (!Number.isFinite(minutes) || minutes < 1) return null
  return Math.floor(minutes)
}

export function readDataExportEnabled(status: unknown): boolean | null {
  const raw = readStatusValue(status, 'enable_data_export')
  return typeof raw === 'boolean' ? raw : null
}

export function readDataExportLastSuccessAt(status: unknown): number | null {
  const raw = readStatusValue(status, 'data_export_last_success_at')
  const timestamp = typeof raw === 'number' ? raw : Number(raw)
  if (!Number.isFinite(timestamp) || timestamp < 1) return null
  return Math.floor(timestamp)
}

export function formatQuotaDataFreshnessMessage(
  t: (key: string, options?: Record<string, unknown>) => string,
  options: { language?: string; status?: unknown }
): string | null {
  const enabled = readDataExportEnabled(options.status)
  if (enabled === null) return null
  if (!enabled) return t('Dashboard data export is disabled.')

  const exportedAt = readDataExportLastSuccessAt(options.status)
  if (!exportedAt) return t('Dashboard data has not been exported yet.')

  const time = formatQuotaDataExportTime(exportedAt * 1000, options.language)
  const minutes = readDataExportIntervalMinutes(options.status)
  if (minutes) {
    return t(
      'Data last exported at {{time}}. Updated about every {{minutes}} minutes.',
      { time, minutes }
    )
  }
  return t('Data last exported at {{time}}.', { time })
}
