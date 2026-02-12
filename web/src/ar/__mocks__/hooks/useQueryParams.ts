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

import React, { useEffect, useMemo, useRef } from 'react'
import { useLocation } from 'react-router-dom'
import qs from 'qs'
import type { IParseOptions } from 'qs'
import { assignWith, isNil } from 'lodash-es'

export interface UseQueryParamsOptions<T> extends IParseOptions {
  /**
   * Keys that should be converted to numbers
   */
  numberKeys?: (keyof T)[]
  /**
   * Keys that should be converted to booleans
   */
  booleanKeys?: (keyof T)[]
  processQueryParams?(data: any): T
}

export function useQueryParams<T = unknown>(options?: UseQueryParamsOptions<T>): T {
  const { search } = useLocation()

  const queryParams = React.useMemo(() => {
    const {
      numberKeys: _n,
      booleanKeys: _b,
      processQueryParams: _p,
      ...qsOptions
    } = (options ?? {}) as UseQueryParamsOptions<T> & Record<string, unknown>
    const parsed = qs.parse(search, { ignoreQueryPrefix: true, ...qsOptions }) as Record<string, any>
    const params = { ...parsed }

    options?.numberKeys?.forEach(key => {
      const value = params[key as string]
      if (typeof value === 'string' && value.trim() !== '' && !isNaN(Number(value))) {
        params[key as string] = Number(value)
      }
    })

    options?.booleanKeys?.forEach(key => {
      const value = params[key as string]
      if (value === 'true') params[key as string] = true
      if (value === 'false') params[key as string] = false
    })

    let result = params as T
    if (typeof options?.processQueryParams === 'function') {
      result = options.processQueryParams(result)
    }
    return result
  }, [search, options])

  return queryParams as T
}

/** Keys that are always kept as strings (e.g. searchTerm to preserve leading zeros) */
const STRING_KEYS = ['searchTerm']

export const useQueryParamsOptions = <Q extends object, DKey extends keyof Q>(
  defaultParams: { [K in DKey]: NonNullable<Q[K]> },
  _decoderOptions?: { ignoreEmptyString?: boolean }
): UseQueryParamsOptions<Required<Pick<Q, DKey>>> => {
  const defaultParamsRef = useRef(defaultParams)
  useEffect(() => {
    defaultParamsRef.current = defaultParams
  }, [defaultParams])

  return useMemo(
    () => ({
      numberKeys: ['page', 'size'] as (keyof Required<Pick<Q, DKey>>)[],
      processQueryParams: (params: Q) => {
        const processed = { ...params }

        STRING_KEYS.forEach(key => {
          if (processed[key as keyof Q] != null) {
            ;(processed as Record<string, unknown>)[key] = String((processed as Record<string, unknown>)[key])
          }
        })

        const withDefaults = assignWith(processed, defaultParamsRef.current, (objValue, srcValue) =>
          isNil(objValue) ? srcValue : objValue
        ) as Record<string, unknown>

        if ('sort' in withDefaults && typeof withDefaults.sort === 'string') {
          withDefaults.sort = [withDefaults.sort]
        }
        return withDefaults as Required<Pick<Q, DKey>>
      }
    }),
    []
  )
}
