# Adding Swift as a New Registry Type – Checklist

Swift is added like existing types (Python, Maven, Generic, etc.): same API usage and UI, with a **Swift-specific icon** only.

---

## 1. New files to create

### 1.1 Version (artifact/version UI)

| File | Purpose |
|------|--------|
| `web/src/ar/pages/version-details/SwiftVersion/SwiftVersionType.tsx` | Version type class (copy from `GenericVersion/GenericVersionType.tsx`, set `packageType = RepositoryPackageType.SWIFT`, icon/tree labels as needed). |
| `web/src/ar/pages/version-details/SwiftVersion/types.ts` | Types for Swift version (copy from `GenericVersion/types.ts` if you need artifact-detail types). |
| `web/src/ar/pages/version-details/SwiftVersion/pages/overview/SwiftOverviewPage.tsx` | Overview tab content (copy from `GenericVersion/pages/overview/OverviewPage.tsx`). |
| `web/src/ar/pages/version-details/SwiftVersion/pages/overview/SwiftVersionGeneralInfo.tsx` | General info card showing package type label (use `getString('packageTypes.swiftPackage')` and Swift icon). |
| `web/src/ar/pages/version-details/SwiftVersion/pages/overview/styles.module.scss` | Styles for overview (copy from Generic). |
| `web/src/ar/pages/version-details/SwiftVersion/pages/overview/styles.module.scss.d.ts` | Typings for overview styles. |
| `web/src/ar/pages/version-details/SwiftVersion/pages/artifact-details/SwiftArtifactDetailsPage.tsx` | Files tab / artifact details page (copy from `GenericVersion/pages/artifact-details/GenericArtifactDetailsPage.tsx`). |
| `web/src/ar/pages/version-details/SwiftVersion/pages/oss-details/OSSContentPage.tsx` | OSS tab (if Swift supports it; copy from Generic). |
| `web/src/ar/pages/version-details/SwiftVersion/pages/oss-details/OSSArtifactDetailsContent.tsx` | OSS artifact content (if needed). |
| `web/src/ar/pages/version-details/SwiftVersion/pages/oss-details/OSSGeneralInfoContent.tsx` | OSS general info (if needed). |
| `web/src/ar/pages/version-details/SwiftVersion/pages/oss-details/ossDetails.module.scss` (+ `.d.ts`) | OSS styles (if needed). |

**Minimal set (same behavior as Generic):**  
`SwiftVersionType.tsx`, `types.ts`, overview page + `SwiftVersionGeneralInfo.tsx` + styles, and `SwiftArtifactDetailsPage.tsx`. Add OSS files only if Swift should have the OSS tab.

### 1.2 Repository (registry setup UI)

| File | Purpose |
|------|--------|
| `web/src/ar/pages/repository-details/SwiftRepository/SwiftRepositoryType.tsx` | Repository type class (copy from `GenericRepository/GenericRepositoryType.tsx`, set `packageType = RepositoryPackageType.SWIFT`, `repositoryIcon = 'swift-icon'` or chosen icon). |

Optional (if you add tests):

- `web/src/ar/pages/repository-details/SwiftRepository/__tests__/__mockData__.ts`
- `web/src/ar/pages/repository-details/SwiftRepository/__tests__/CreateSwiftRegistry.test.tsx`
- `web/src/ar/pages/version-details/SwiftVersion/__tests__/` (e.g. list/overview/artifact details tests, copy from Generic).

---

## 2. Existing files to modify

### 2.1 Types and enums

| File | Change |
|------|--------|
| `web/src/ar/common/types.ts` | In `RepositoryPackageType` enum add: `SWIFT = 'SWIFT'`. |

### 2.2 Upstream proxy (if Swift supports upstream)

| File | Change |
|------|--------|
| `web/src/ar/pages/upstream-proxy-details/types.ts` | In `UpstreamProxyPackageType` enum add: `SWIFT = 'SWIFT'`. |

### 2.3 Factories (registration)

| File | Change |
|------|--------|
| `web/src/ar/pages/version-details/VersionFactory.tsx` | Import `SwiftVersionType`, then `versionFactory.registerStep(new SwiftVersionType())`. |
| `web/src/ar/pages/repository-details/RepositoryFactory.tsx` | Import `SwiftRepositoryType`, then `repositoryFactory.registerStep(new SwiftRepositoryType())`. |

### 2.4 Repository/package type lists (dropdowns and artifact list “type” column)

| File | Change |
|------|--------|
| `web/src/ar/hooks/useGetRepositoryTypes.ts` | In `RepositoryTypes` array add: `{ label: 'repositoryTypes.swift', value: RepositoryPackageType.SWIFT, icon: 'swift-icon' }` (or chosen icon name). |
| `web/src/ar/hooks/useGetUpstreamRepositoryPackageTypes.ts` | In `UpstreamProxyPackageTypeList` add same entry with `value: UpstreamProxyPackageType.SWIFT` (if Swift supports upstream). |

### 2.5 Strings (labels and types)

| File | Change |
|------|--------|
| `web/src/ar/strings/strings.en.yaml` | Under `packageTypes:` add `swiftPackage: Swift Package`. Under `repositoryTypes:` add `swift: Swift`. |
| `web/src/ar/strings/types.ts` | In `StringsMap` add `'packageTypes.swiftPackage': string` and `'repositoryTypes.swift': string`. |

### 2.6 Version type implementation details

- In **`SwiftVersionType.tsx`**:
  - Extend the same base as Generic (e.g. `VersionStep<ArtifactVersionSummary>`).
  - Set `protected packageType = RepositoryPackageType.SWIFT`.
  - Set `protected hasArtifactRowSubComponent = true` if the artifact list row should expand (e.g. to show files).
  - Point overview tab to `SwiftOverviewPage` and general info to `SwiftVersionGeneralInfo`.
  - Point artifact-details tab to a page that uses `VersionFilesProvider` + `ArtifactFilesContent` (or `SwiftArtifactDetailsPage` which does the same).
  - Use `renderArtifactTreeNodeView` / `renderVersionTreeNodeView` with your chosen **Swift icon** (e.g. `'swift-icon'`).
- In **`SwiftVersionGeneralInfo.tsx`**: use `getString('packageTypes.swiftPackage')` and the same Swift icon for the package type row.

### 2.7 Repository type implementation details

- In **`SwiftRepositoryType.tsx`**:
  - Set `protected packageType = RepositoryPackageType.SWIFT`.
  - Set `protected repositoryIcon: IconName = 'swift-icon'` (or the icon you use in `@harnessio/icons`).
  - Keep the same create/configuration/actions/setup client/redirect/tree methods as Generic (or as needed).

### 2.8 Icon

- **Icon name:** Use one of:
  - An existing icon in `@harnessio/icons` that fits Swift (e.g. a code/package icon), or
  - A new icon (e.g. `swift-icon`) if your design system adds it.
- Replace every `'swift-icon'` in this checklist with the actual icon name you use.

---

## 3. Backend / API

- Ensure the backend and OpenAPI client accept **`SWIFT`** as a valid `package_type` (or equivalent) for:
  - Creating/updating registries
  - Listing/filtering artifacts and versions
  - Version details and file listing
- If the backend uses an enum, add `SWIFT` there and regenerate the client if needed.

---

## 4. Summary table

| Category | New files | Modified files |
|----------|-----------|----------------|
| Version (SwiftVersion) | `SwiftVersionType.tsx`, `types.ts`, overview page + GeneralInfo + styles, `SwiftArtifactDetailsPage.tsx`, (optional) OSS pages + styles | — |
| Repository (SwiftRepository) | `SwiftRepositoryType.tsx` | — |
| Registration & types | — | `common/types.ts`, `upstream-proxy-details/types.ts`, `VersionFactory.tsx`, `RepositoryFactory.tsx` |
| Lists & strings | — | `useGetRepositoryTypes.ts`, `useGetUpstreamRepositoryPackageTypes.ts`, `strings.en.yaml`, `strings/types.ts` |

No changes are required in **routes** or **RouteDestinations**: version/repository resolution is done via `versionFactory.getVersionType(packageType)` and `repositoryFactory.getRepositoryType(packageType)` using the new `SWIFT` enum value and registered types.
