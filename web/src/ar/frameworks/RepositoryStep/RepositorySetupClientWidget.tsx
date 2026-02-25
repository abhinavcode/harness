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

import React from 'react'
import { Text } from '@harnessio/uicore'

import { useStrings } from '@ar/frameworks/strings'
import type { RepositoryPackageType } from '@ar/common/types'

import repositoryFactory from './RepositoryFactory'
import type { RepositoySetupClientProps } from './Repository'
import type { RepositoryAbstractFactory } from './RepositoryAbstractFactory'

interface RepositorySetupClientWidgetProps extends RepositoySetupClientProps {
  factory?: RepositoryAbstractFactory
  type: RepositoryPackageType
  /** When true (e.g. from registry list), use registryPath for client-setup-details API. Injected here so types don't need to prop-drill. */
  isSentFromRegistryPage?: boolean
  /** Full path from list API (accountId/orgId/projectId/registryName). Injected here when isSentFromRegistryPage is true. */
  registryPath?: string
}

export default function RepositorySetupClientWidget(props: RepositorySetupClientWidgetProps): JSX.Element {
  const {
    factory = repositoryFactory,
    type,
    onClose,
    repoKey,
    artifactKey,
    versionKey,
    isSentFromRegistryPage,
    registryPath
  } = props
  const { getString } = useStrings()
  const repositoryType = factory?.getRepositoryType(type)
  if (!repositoryType) {
    return <Text intent="warning">{getString('stepNotFound')}</Text>
  }
  const content = repositoryType.renderSetupClient({
    onClose,
    repoKey,
    artifactKey,
    versionKey
  })
  if (!content || !React.isValidElement(content)) {
    return content
  }
  return React.cloneElement(content, {
    isSentFromRegistryPage,
    registryPath
  } as React.Attributes & { isSentFromRegistryPage?: boolean; registryPath?: string })
}
