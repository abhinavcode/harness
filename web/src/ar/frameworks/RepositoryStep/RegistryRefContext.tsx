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

import React, { createContext, useContext } from 'react'

/**
 * Pre-computed registry ref (e.g. from list path). Set by RepositoryActionsWidget when
 * opened from the registry list; consumed by RepositoryActions/UpstreamProxyActions for
 * the Setup Client flow. When not set (e.g. on details page), useRegistryRef() returns
 * undefined and SetupClientContent falls back to scope-based ref.
 */
const RegistryRefContext = createContext<string | undefined>(undefined)

export function RegistryRefProvider({
  value,
  children
}: {
  value: string | undefined
  children: React.ReactNode
}): React.ReactElement {
  return <RegistryRefContext.Provider value={value}>{children}</RegistryRefContext.Provider>
}

export function useRegistryRef(): string | undefined {
  return useContext(RegistryRefContext)
}
