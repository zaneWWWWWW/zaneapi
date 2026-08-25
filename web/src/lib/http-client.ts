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
import axios, { type AxiosRequestConfig } from 'axios'
import { t } from 'i18next'
import { toast } from 'sonner'

import {
  applyAuthRotation,
  clearAuthentication,
  refreshAuthentication,
} from '@/lib/auth-session'
import { getServerErrorMessageKey } from '@/lib/server-error-message'
import { useAuthStore } from '@/stores/auth-store'

declare module 'axios' {
  export interface AxiosRequestConfig {
    skipBusinessError?: boolean
    skipErrorHandler?: boolean
    disableDuplicate?: boolean
    skipAuthRefresh?: boolean
    authRetry?: boolean
    acceptAuthRotation?: boolean
  }
}

export type ApiRequestConfig = AxiosRequestConfig

export const api = axios.create({
  baseURL: '',
  withCredentials: true,
  headers: {
    'Cache-Control': 'no-store',
  },
})

const inFlightGet = new Map<string, Promise<unknown>>()
const originalGet = api.get.bind(api)

api.get = ((url: string, config: ApiRequestConfig = {}) => {
  if (config.disableDuplicate) return originalGet(url, config)

  const params = config.params ? JSON.stringify(config.params) : '{}'
  const sessionSID = useAuthStore.getState().auth.session?.sid || 'anonymous'
  const key = `${sessionSID}:${url}?${params}`
  const existingRequest = inFlightGet.get(key)
  if (existingRequest) return existingRequest

  const request = originalGet(url, config).finally(() => {
    inFlightGet.delete(key)
  })
  inFlightGet.set(key, request)
  return request
}) as typeof api.get

const PUBLIC_PATHS_AFTER_AUTH_LOSS = [
  '/',
  '/sign-in',
  '/sign-up',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/oauth',
  '/privacy-policy',
  '/user-agreement',
  '/about',
  '/docs',
  '/setup',
  '/pricing',
  '/rankings',
]

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object'
}

function translateToast(key: string): string {
  const value = t(key)
  return typeof value === 'string' && value.trim() ? value : key
}

export function resolveHttpErrorToastMessage(error: {
  response?: { status?: number; data?: unknown }
  message?: string
}): string {
  const messageKey = getServerErrorMessageKey(error)
  if (messageKey) return translateToast(messageKey)

  const status = error.response?.status
  if (status === 429) {
    return translateToast('Too many requests')
  }

  const data = error.response?.data
  if (isRecord(data) && typeof data.message === 'string' && data.message.trim()) {
    return data.message
  }

  const fallback = error.message?.trim() ?? ''
  if (fallback && !/^Request failed with status code \d+$/.test(fallback)) {
    return fallback
  }

  return translateToast('Request failed')
}

export function isPublicPathAfterAuthLoss(pathname: string): boolean {
  const path = pathname.split('?')[0]?.split('#')[0] || '/'
  if (path === '/') return true
  return PUBLIC_PATHS_AFTER_AUTH_LOSS.some(
    (prefix) => prefix !== '/' && (path === prefix || path.startsWith(`${prefix}/`))
  )
}

function handleLostSession(hadSession: boolean, skipErrorHandler?: boolean): void {
  if (!hadSession) return
  if (
    typeof window !== 'undefined' &&
    isPublicPathAfterAuthLoss(window.location.pathname)
  ) {
    return
  }
  if (!skipErrorHandler) toast.error(t('Session expired!'))
  if (typeof window !== 'undefined' && window.location.pathname !== '/') {
    window.location.replace('/')
  }
}

api.interceptors.response.use(
  (response) => {
    if (response.config.acceptAuthRotation && response.data?.success === true) {
      applyAuthRotation(response.data.data)
    }

    if (
      !response.config.skipBusinessError &&
      typeof response.data?.success === 'boolean' &&
      !response.data.success
    ) {
      const messageKey = getServerErrorMessageKey(response.data)
      toast.error(
        messageKey
          ? t(messageKey)
          : response.data.message || t('Request failed')
      )
    }
    return response
  },
  async (error) => {
    const config = error?.config as ApiRequestConfig | undefined
    const skipErrorHandler = config?.skipErrorHandler
    const status = error?.response?.status

    if (status === 401) {
      const hadSession = Boolean(useAuthStore.getState().auth.accessToken)
      if (config && !config.skipAuthRefresh && !config.authRetry) {
        config.authRetry = true
        const outcome = await refreshAuthentication()
        if (outcome.kind === 'authenticated') {
          const token = useAuthStore.getState().auth.accessToken
          if (token) {
            config.headers = {
              ...config.headers,
              Authorization: `Bearer ${token}`,
            }
          }
          return api.request(config)
        }

        if (outcome.kind === 'anonymous' || outcome.kind === 'out_of_sync') {
          handleLostSession(hadSession, skipErrorHandler)
        }
      } else if (config?.authRetry) {
        clearAuthentication(false)
        handleLostSession(hadSession, skipErrorHandler)
      } else if (!skipErrorHandler && hadSession) {
        toast.error(t('Session expired!'))
      }
    } else if (!skipErrorHandler) {
      toast.error(resolveHttpErrorToastMessage(error))
    }
    throw error
  }
)

api.interceptors.request.use((config) => {
  const accessToken = useAuthStore.getState().auth.accessToken
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }
  return config
})
