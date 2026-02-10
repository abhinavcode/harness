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

import React, { useContext, useRef } from 'react'
import { Container, Layout, ExpandingSearchInput, ExpandingSearchInputHandle, PageError } from '@harnessio/uicore'
import { Icon } from '@harnessio/icons'

import { DEFAULT_PAGE_INDEX } from '@ar/constants'
import { useStrings } from '@ar/frameworks/strings'
import { VersionFilesContext } from '@ar/pages/version-details/context/VersionFilesProvider'
import ArtifactFileListTable from '@ar/pages/version-details/components/ArtifactFileListTable/ArtifactFileListTable'

import css from './ArtifactFilesContent.module.scss'

interface ArtifactFilesContentProps {
  minimal?: boolean
}
export default function ArtifactFilesContent(props: ArtifactFilesContentProps): JSX.Element {
  const { minimal } = props
  const { getString } = useStrings()
  const searchRef = useRef<ExpandingSearchInputHandle>({} as ExpandingSearchInputHandle)
  const { data, loading, error, refetch, updateQueryParams, sort, queryParams } = useContext(VersionFilesContext)

  return (
    <Layout.Vertical spacing="medium">
      {!minimal && (
        <ExpandingSearchInput
          alwaysExpanded
          width={500}
          placeholder={getString('search')}
          onChange={text => {
            updateQueryParams({ searchTerm: text || undefined, page: DEFAULT_PAGE_INDEX })
          }}
          defaultValue={queryParams?.searchTerm || ''}
          ref={searchRef}
        />
      )}
      <Container className={css.tableContainer}>
        {loading ? (
          <Container flex={{ justifyContent: 'center', alignItems: 'center' }} padding="xlarge">
            <Icon name="steps-spinner" size={32} />
          </Container>
        ) : error && !loading ? (
          <PageError
            message={
              error instanceof Error
                ? error.message
                : typeof error === 'object' && error !== null && 'message' in error
                ? String((error as { message?: unknown }).message)
                : String(error)
            }
            onClick={() => refetch()}
          />
        ) : data ? (
          <ArtifactFileListTable
            data={data}
            gotoPage={pageNumber => updateQueryParams({ page: pageNumber })}
            setSortBy={sortArr => {
              updateQueryParams({ sort: sortArr, page: DEFAULT_PAGE_INDEX })
            }}
            sortBy={sort}
            minimal={minimal}
          />
        ) : null}
      </Container>
    </Layout.Vertical>
  )
}
