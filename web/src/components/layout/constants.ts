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
/**
 * Layout constants and configurations
 */

export const DEFAULT_APP_LAYOUT = {
  variant: 'sidebar',
  collapsible: 'icon',
} as const

export const APP_SIDEBAR_METRICS = {
  desktopWidth: '16.25rem',
  mobileWidth: '17rem',
  collapsedWidth: '3.25rem',
} as const

/**
 * ChatGPT-style chrome: the top bar is a utility strip, not a sitemap.
 * Destination links live in the sidebar (workspace) or page body / footer
 * (public). The header only carries brand + account actions.
 */
export const APP_HEADER_CHROME = {
  workspaceNavLinks: false,
  publicNavLinks: false,
  workspaceSearch: 'none',
} as const

export type AppLayoutVariant = 'sidebar' | 'inset' | 'floating'

/**
 * Restore a saved layout variant, or fall back to the connected sidebar.
 * Unknown / empty cookies must not resurrect the old floating `inset` default.
 */
export function resolveAppLayoutVariant(
  saved: string | null | undefined
): AppLayoutVariant {
  if (saved === 'sidebar' || saved === 'inset' || saved === 'floating') {
    return saved
  }
  return DEFAULT_APP_LAYOUT.variant
}

/**
 * Animation variants for mobile drawer
 */
export const MOBILE_DRAWER_ANIMATION = {
  overlay: {
    hidden: { opacity: 0 },
    visible: { opacity: 1 },
    exit: { opacity: 0 },
  },
  drawer: {
    hidden: { opacity: 0, y: 16 },
    visible: {
      opacity: 1,
      y: 0,
      transition: {
        duration: 0.2,
        ease: [0.16, 1, 0.3, 1],
      },
    },
    exit: {
      opacity: 0,
      y: 16,
      transition: { duration: 0.15 },
    },
  },
  menuItem: {
    hidden: { opacity: 0 },
    visible: { opacity: 1 },
  },
} as const

/**
 * Mobile drawer configuration
 */
export const MOBILE_DRAWER_CONFIG = {
  overlayTransitionDuration: 0.2,
  drawerClassName:
    'fixed inset-x-0 bottom-3 z-50 mx-auto w-[95%] rounded-lg border border-border bg-background p-4 shadow-lg md:hidden',
  overlayClassName: 'fixed inset-0 z-40 bg-black/50',
} as const
