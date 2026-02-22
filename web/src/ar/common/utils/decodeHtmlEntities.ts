/*
 * Copyright 2024 Harness, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * Safely decodes HTML entities in a string without XSS vulnerability.
 * Uses DOMParser which is safer than innerHTML as it doesn't execute scripts.
 *
 * @param text - The text containing HTML entities to decode
 * @returns The decoded text with HTML entities converted to their character equivalents
 *
 * @example
 * decodeHtmlEntities('github.com&#x2F;rs&#x2F;xid') // returns 'github.com/rs/xid'
 * decodeHtmlEntities('&lt;script&gt;alert(1)&lt;/script&gt;') // returns '<script>alert(1)</script>' (safe, no execution)
 */
export const decodeHtmlEntities = (text: string): string => {
  if (!text || typeof text !== 'string') {
    return text
  }

  // Use DOMParser which is safer than innerHTML
  // It parses the content as text/html but doesn't execute scripts
  const parser = new DOMParser()
  const doc = parser.parseFromString(text, 'text/html')

  // Extract the text content which automatically decodes entities
  return doc.documentElement.textContent || text
}
