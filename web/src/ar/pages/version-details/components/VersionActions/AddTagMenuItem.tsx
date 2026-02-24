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

import React, { useState } from 'react'
import { Formik } from 'formik'
import { Button, ButtonVariation, Layout, ModalDialog, useToaster } from '@harnessio/uicore'

import { useAppStore, useParentComponents } from '@ar/hooks'
import { useParentUtils } from '@ar/hooks/useParentUtils'
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
  const { getApiBaseUrl, getCustomHeaders } = useParentUtils()
  const accountId = typeof scope?.accountId === 'string' ? scope.accountId : ''
  const { getString } = useStrings()
  const { showSuccess, showError } = useToaster()
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (tagNamesRaw: string[] | (string | { label: string; value: string })[]) => {
    const trimmed = normalizeTagNames(tagNamesRaw)
    if (trimmed.length === 0) {
      showError(getString('validationMessages.entityRequired', { entity: 'Tag name' }))
      return
    }
    setLoading(true)
    try {
      const path = '/har/api/v2/oci/tags'
      const baseUrl = new URL(getApiBaseUrl(''), window.location.origin)
      console.log('baseUrl', baseUrl);
      baseUrl.pathname = path
      console.log('baseUrl after adding path', baseUrl);
      baseUrl.searchParams.set('account_identifier', accountId || '')
      baseUrl.searchParams.set('registry_identifier', repoKey)
      const url = baseUrl.toString()
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...getCustomHeaders()
      }
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 30000) // 30 second timeout
      let res: Response
      try {
        res = await fetch(url, {
          method: 'POST',
          headers,
          body: JSON.stringify({
            package: artifactKey,
            version: versionKey,
            tags: trimmed
          }),
          signal: controller.signal
        })
        clearTimeout(timeoutId)
      } catch (fetchError) {
        clearTimeout(timeoutId)
        throw fetchError
      }
      if (!res.ok) {
        const errText = await res.text()
        throw new Error(errText || res.statusText)
      }
      showSuccess(getString('versionList.messages.addTagSuccess'))
      hideModal()
      onClose?.()
      queryClient.invalidateQueries(['GetAllHarnessArtifacts'])
      queryClient.invalidateQueries(['ListVersions'])
    } catch (e) {
      showError((e as Error)?.message || getString('versionList.messages.addTagFailed'))
    } finally {
      setLoading(false)
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
                onClick={() => handleSubmit(formik.values.tagNames)}
                disabled={loading}>
                {getString('add')}
              </Button>
              <Button variation={ButtonVariation.SECONDARY} onClick={handleClose} disabled={loading}>
                {getString('cancel')}
              </Button>
            </Layout.Horizontal>
          }>
          <div className={css.addTagInputWrapper}>
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
