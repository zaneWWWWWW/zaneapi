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
import i18next from 'i18next'
import { useCallback, useState } from 'react'
import { toast } from 'sonner'

import { isApiSuccess, requestHupijiaoPayment } from '../api'

function getPaymentURL(data: unknown): string | null {
  if (!data || typeof data !== 'object') return null
  if ('url' in data && typeof data.url === 'string') return data.url
  return null
}

function isSafePaymentURL(value: string): boolean {
  try {
    const url = new URL(value.trim())
    return (
      url.protocol === 'https:' &&
      !url.username &&
      !url.password &&
      Boolean(url.hostname)
    )
  } catch {
    return false
  }
}

export function useHupijiaoPayment() {
  const [processing, setProcessing] = useState(false)

  const processHupijiaoPayment = useCallback(async (topupAmount: number) => {
    setProcessing(true)
    try {
      const response = await requestHupijiaoPayment({
        amount: Math.floor(topupAmount),
      })
      const paymentURL = isApiSuccess(response)
        ? getPaymentURL(response.data)
        : null

      if (paymentURL) {
        if (!isSafePaymentURL(paymentURL)) {
          toast.error(i18next.t('Invalid payment redirect URL'))
          return false
        }
        toast.success(i18next.t('Redirecting to payment page...'))
        window.location.href = paymentURL
        return true
      }

      toast.error(response.message || i18next.t('Payment request failed'))
      return false
    } catch {
      toast.error(i18next.t('Payment request failed'))
      return false
    } finally {
      setProcessing(false)
    }
  }, [])

  return { processing, processHupijiaoPayment }
}
