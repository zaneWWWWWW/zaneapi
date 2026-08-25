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
import { zodResolver } from '@hookform/resolvers/zod'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import * as z from 'zod'

import { Checkbox } from '@/components/ui/checkbox'
import {
  Form,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { parseGroupAvailabilityGroups } from '@/features/dashboard/lib/group-availability'
import { getGroups } from '@/features/users/api'

import { SettingsForm } from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

const groupAvailabilitySchema = z.object({
  groupsJson: z.string(),
})

type GroupAvailabilityFormValues = z.infer<typeof groupAvailabilitySchema>

type GroupAvailabilitySectionProps = {
  groupsJson: string
}

export function GroupAvailabilitySection(props: GroupAvailabilitySectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const groupsQuery = useQuery({
    queryKey: ['groups'],
    queryFn: getGroups,
  })

  const form = useForm<GroupAvailabilityFormValues>({
    resolver: zodResolver(groupAvailabilitySchema),
    defaultValues: {
      groupsJson: props.groupsJson || '[]',
    },
  })

  useEffect(() => {
    form.reset({ groupsJson: props.groupsJson || '[]' })
  }, [form, props.groupsJson])

  const groupsJson = form.watch('groupsJson')
  const availableGroups = useMemo(() => {
    const fromApi = groupsQuery.data?.data ?? []
    const selected = parseGroupAvailabilityGroups(groupsJson)
    const names = [...fromApi]
    if (!names.includes('auto')) names.push('auto')
    for (const name of selected) {
      if (!names.includes(name)) names.push(name)
    }
    return names
  }, [groupsJson, groupsQuery.data?.data])

  const onSubmit = async (values: GroupAvailabilityFormValues) => {
    const nextValue = values.groupsJson || '[]'
    const previousValue = props.groupsJson || '[]'
    if (nextValue === previousValue) return
    await updateOption.mutateAsync({
      key: 'console_setting.group_availability_groups',
      value: nextValue,
    })
  }

  return (
    <SettingsSection title={t('Group availability')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)}>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
          />
          <FormField
            control={form.control}
            name='groupsJson'
            render={({ field }) => {
              const selected = parseGroupAvailabilityGroups(field.value)
              return (
                <FormItem>
                  <FormLabel>{t('Groups shown on the overview')}</FormLabel>
                  <FormDescription>
                    {t(
                      'Choose which groups appear in dashboard group availability. Leave all unchecked to hide the panel.'
                    )}
                  </FormDescription>
                  {availableGroups.length === 0 ? (
                    <p className='text-muted-foreground text-sm'>
                      {t('No groups configured')}
                    </p>
                  ) : (
                    <div className='mt-3 grid gap-2 sm:grid-cols-2'>
                      {availableGroups.map((name) => {
                        const checked = selected.includes(name)
                        return (
                          <label
                            key={name}
                            className='border-border hover:bg-muted/30 flex cursor-pointer items-center gap-3 rounded-lg border px-3 py-2.5'
                          >
                            <Checkbox
                              checked={checked}
                              onCheckedChange={(nextChecked) => {
                                const enabled = nextChecked === true
                                let next = selected
                                if (enabled && !checked) {
                                  next = [...selected, name]
                                } else if (!enabled && checked) {
                                  next = selected.filter(
                                    (groupName) => groupName !== name
                                  )
                                }
                                field.onChange(JSON.stringify(next))
                              }}
                            />
                            <span className='font-mono text-sm font-medium'>
                              {name}
                            </span>
                          </label>
                        )
                      })}
                    </div>
                  )}
                  <FormMessage />
                </FormItem>
              )
            }}
          />
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
