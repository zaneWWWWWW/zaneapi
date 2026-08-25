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
import type { UptimeGroupResult, UptimeMonitor } from '../types'

export function normalizeProbeName(value: string): string {
  return value.trim().replaceAll(/\s+/g, ' ').toLowerCase()
}

export function flattenUptimeMonitors(
  groups: UptimeGroupResult[] | null | undefined
): UptimeMonitor[] {
  if (!groups?.length) return []
  const monitors: UptimeMonitor[] = []
  for (const group of groups) {
    for (const monitor of group.monitors ?? []) {
      monitors.push(monitor)
    }
  }
  return monitors
}

export function buildProbeLookup(
  groups: UptimeGroupResult[] | null | undefined
): Map<string, UptimeMonitor> {
  const lookup = new Map<string, UptimeMonitor>()
  for (const monitor of flattenUptimeMonitors(groups)) {
    const key = normalizeProbeName(monitor.name)
    if (key && !lookup.has(key)) lookup.set(key, monitor)
  }
  return lookup
}

export function matchChannelProbe(
  channelName: string,
  lookup: Map<string, UptimeMonitor>
): UptimeMonitor | null {
  return lookup.get(normalizeProbeName(channelName)) ?? null
}

export function listUnmatchedMonitors(
  groups: UptimeGroupResult[] | null | undefined,
  channelNames: string[]
): UptimeMonitor[] {
  const lookup = buildProbeLookup(groups)
  const used = new Set(channelNames.map(normalizeProbeName))
  const unmatched: UptimeMonitor[] = []
  for (const [key, monitor] of lookup) {
    if (!used.has(key)) unmatched.push(monitor)
  }
  return unmatched
}

export type ProbeSummary = {
  up: number
  down: number
  other: number
  total: number
}

export function summarizeProbes(monitors: UptimeMonitor[]): ProbeSummary {
  const summary: ProbeSummary = { up: 0, down: 0, other: 0, total: monitors.length }
  for (const monitor of monitors) {
    if (monitor.status === 1) summary.up += 1
    else if (monitor.status === 0) summary.down += 1
    else summary.other += 1
  }
  return summary
}
