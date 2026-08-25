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
import { RichContent } from '@/components/rich-content'
import { isHttpUrl, isLikelyHtml } from '@/lib/content-format'

type SiteDocumentProps = {
  content: string
  title?: string
  variant?: 'main' | 'section'
}

export function SiteDocument(props: SiteDocumentProps) {
  const content = props.content.trim()
  if (!content) return null

  const isUrl = isHttpUrl(content)
  const contentIsHtml = isLikelyHtml(content)
  const isMain = props.variant !== 'section'

  return (
    <section className='space-y-3'>
      {props.title ? (
        <h2 className='text-lg font-medium'>{props.title}</h2>
      ) : null}
      {isUrl ? (
        <iframe
          src={content}
          className={
            isMain
              ? 'h-[min(40rem,75vh)] w-full rounded-lg border'
              : 'h-[min(32rem,70vh)] w-full rounded-lg border'
          }
          title={props.title}
          sandbox='allow-forms allow-popups allow-popups-to-escape-sandbox allow-scripts'
        />
      ) : (
        <div className={isMain ? undefined : 'rounded-lg border px-4 py-4'}>
          <RichContent
            mode={contentIsHtml ? 'html' : 'markdown'}
            htmlVariant='isolated'
            content={content}
            className='prose-neutral dark:prose-invert max-w-none'
          />
        </div>
      )}
    </section>
  )
}
