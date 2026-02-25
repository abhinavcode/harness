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

import React, { useRef } from 'react'
import { Formik } from 'formik'
import { Button, ButtonVariation, Layout, ModalDialog, useToaster } from '@harnessio/uicore'
import { useAddOciArtifactTagsMutation } from '@harnessio/react-har-service-client'

import { useAppStore, useParentComponents } from '@ar/hooks'
import { useStrings } from '@ar/frameworks/strings'
import { queryClient } from '@ar/utils/queryClient'
import { ResourceType } from '@ar/common/permissionTypes'
import { PermissionIdentifier } from '@ar/common/permissionTypes'
import PatternInput from '@ar/components/Form/PatternInput/PatternInput'

import type { VersionActionProps } from './types'
import css from './AddTagMenuItem.module.scss'

function normalizeTagNames(value: string[] | (string | { label: string; value: string })[]): string[] {
  if (!Array.isArray(value)) return []
  return value
    .map(item => (typeof item === 'string' ? item : item?.value ?? ''))
    .map(t => t.trim())
    .filter(Boolean)
}

export interface AddTagModalContentProps {
  artifactKey: string
  repoKey: string
  versionKey: string
  hideModal: () => void
  onClose?: () => void
}

export function AddTagModalContent({
  artifactKey,
  repoKey,
  versionKey,
  hideModal,
  onClose
}: AddTagModalContentProps): JSX.Element {
  const { scope } = useAppStore()
  const { getString } = useStrings()
  const { showSuccess, showError } = useToaster()
  const { mutateAsync: addOciArtifactTags, isLoading: loading } = useAddOciArtifactTagsMutation()
  const inputWrapperRef = useRef<HTMLDivElement>(null)

  const handleSubmit = async (
    tagNamesRaw: string[] | (string | { label: string; value: string })[],
    currentInputValue?: string
  ) => {
    const committed = normalizeTagNames(tagNamesRaw)
    const current = currentInputValue?.trim() ?? ''
    const allTags = current ? [...committed, current] : committed
    console.log('allTags', allTags);
    const trimmed = normalizeTagNames(allTags)
    if (trimmed.length === 0) {
      showError(getString('validationMessages.entityRequired', { entity: 'Tag name' }))
      return
    }
    try {
      await addOciArtifactTags({
        queryParams: {
          account_identifier: typeof scope?.accountId === 'string' ? scope.accountId : '',
          org_identifier: typeof scope?.orgId === 'string' ? scope.orgId : undefined,
          project_identifier: typeof scope?.projectId === 'string' ? scope.projectId : undefined,
          registry_identifier: repoKey
        },
        body: {
          package: artifactKey,
          version: versionKey,
          tags: trimmed
        }
      })
      showSuccess(getString('versionList.messages.addTagSuccess'))
      hideModal()
      onClose?.()
      queryClient.invalidateQueries(['GetAllHarnessArtifacts'])
      queryClient.invalidateQueries(['ListVersions'])
    } catch (e) {
      showError((e as Error)?.message || getString('versionList.messages.addTagFailed'))
    }
  }

  const handleClose = () => {
    hideModal()
    onClose?.()
  }

  return (
    <Formik initialValues={{ tagNames: [] as string[] }} enableReinitialize onSubmit={() => undefined}>
      {formik => (
        <ModalDialog
          title={getString('versionList.actions.addTag')}
          isOpen={true}
          onClose={handleClose}
          showOverlay={loading}
          footer={
            <Layout.Horizontal spacing="large">
              <Button
                variation={ButtonVariation.PRIMARY}
                onClick={() => {
                  const currentInput =
                    inputWrapperRef.current?.querySelector('input')?.value?.trim() ?? ''
                  handleSubmit(formik.values.tagNames, currentInput)
                }}
                disabled={loading}>
                {getString('add')}
              </Button>
              <Button variation={ButtonVariation.SECONDARY} onClick={handleClose} disabled={loading}>
                {getString('cancel')}
              </Button>
            </Layout.Horizontal>
          }>
          <div ref={inputWrapperRef} className={css.addTagInputWrapper}>
            <PatternInput
              name="tagNames"
              label={getString('versionList.table.columns.tags')}
              placeholder={getString('versionList.actions.addTagPlaceholder')}
              disabled={loading}
            />
          </div>
        </ModalDialog>
      )}
    </Formik>
  )
}

export interface AddTagMenuItemProps extends VersionActionProps {
  openAddTagModal?: () => void
}

export default function AddTagMenuItem(props: AddTagMenuItemProps): JSX.Element {
  const { repoKey, readonly, onClose, openAddTagModal } = props
  const { getString } = useStrings()
  const { RbacMenuItem } = useParentComponents()

  const handleClick = () => {
    openAddTagModal?.()
    onClose?.()
  }

  return (
    <RbacMenuItem
      icon="plus"
      text={getString('versionList.actions.addTag')}
      onClick={handleClick}
      disabled={readonly}
      permission={{
        resource: {
          resourceType: ResourceType.ARTIFACT_REGISTRY,
          resourceIdentifier: repoKey
        },
        permission: PermissionIdentifier.UPLOAD_ARTIFACT
      }}
    />
  )
}
