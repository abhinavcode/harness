# Artifact Registry (AR) – Complete Repo Structure & UI Reference

This document describes the **entire structure** of the AR (Artifact Registry) frontend with **absolute detail**, mapped to the UI. Use it as a single reference for navigation, routing, pages, factories, and data flow.

---

## 1. Overview

- **What AR is**: The Artifact Registry module lets users manage **registries** (repositories), **packages/artifacts**, and **versions** for multiple package types (Docker, NPM, Maven, Swift, Go, etc.).
- **Tech**: React, TypeScript, React Router, TanStack Query, Harness UI Core, YAML strings for i18n.
- **Location**: `web/src/ar/` (all paths below are relative to this unless stated).

---

## 2. Entry Point & App Bootstrap

### 2.1 App entry

- **File**: `app/App.tsx`
- **Role**: Root component of the AR micro-frontend. It:
  - Wraps the app in **QueryClientProvider** (TanStack React Query).
  - Provides **AppStoreContext** (scope, baseUrl, matchPath, parent OSS/Enterprise, repository list view type, public access flags).
  - Provides **StringsContextProvider** with `strings.en.yaml`.
  - Provides **ParentProvider** with hooks and components from the parent (Harness/Gitness) – e.g. `useQueryParams`, `useUpdateQueryParams`, `usePermission`, `RbacButton`, `ModalProvider`, `PageNotPublic`.
  - Registers **factories** by importing:
    - `@ar/pages/version-details/VersionFactory` – registers all version types (Docker, NPM, Swift, …).
    - `@ar/pages/repository-details/RepositoryFactory` – registers all repository types.
  - Renders **NavComponent** (or custom) and inside it **AsyncDownloadRequestsProvider** and lazy **RouteDestinations** (the router).

**UI**: No direct UI; it’s the shell. The first visible UI is either the **Repository List** (list or tree) or a **Redirect** page, depending on the route.

### 2.2 API client

- **File**: `app/useOpenApiClient.ts`
- **Role**: Configures OpenAPI clients (HAR Service, SSCA Manager, NG Manager) with interceptors (e.g. 401 handling) and custom headers from parent. Called once in `App.tsx`.

---

## 3. Routing & URL Structure (UI Mapping)

### 3.1 Route definitions

- **File**: `routes/RouteDefinitions.ts`
- **Path types**: `routes/types.ts`

Routes are built by `routeDefinitions` and can be wrapped with a “mode” prefix (e.g. org/project) via `routeDefinitionWithMode` in `routes/utils.ts`. **Hook**: `useRoutes()` (and `useRoutes(true)` for path-only definitions) from `hooks/useRoutes.ts`.

| Route method | Typical path (relative) | UI |
|--------------|--------------------------|-----|
| `toAR()` | `/` | Redirects to repository list. |
| `toARRedirect()` | `/redirect?...` | Redirect page (e.g. deep links with packageType, registryId, artifactId, versionId). |
| `toARRepositories()` | `/registries` | **Repository list** (list or tree depending on `repositoryListViewType`). |
| `toARRepositoryDetails({ repositoryIdentifier })` | `/registries/:repositoryIdentifier` | **Repository details** – default tab (e.g. Packages). |
| `toARRepositoryDetailsTab({ repositoryIdentifier, tab })` | `/registries/:repositoryIdentifier/:tab` | **Repository details** – specific tab (packages, configuration, webhooks, metadata). |
| `toARArtifacts()` | (context-dependent) | **Global artifact list** (all artifacts across registries). |
| `toARArtifactDetails({ repositoryIdentifier, artifactType, artifactIdentifier })` | `/registries/:repo/:artifactType/:artifactIdentifier` | **Artifact details** – default sub-route (e.g. Versions). |
| `toARArtifactVersions(...)` | `.../versions` | **Artifact details** – **Versions** tab (version list table). |
| `toARArtifactProperties(...)` | `.../properties` | **Artifact details** – **Metadata/Properties** tab. |
| `toARVersionDetails({ ... params, versionIdentifier })` | `.../versions/:versionIdentifier` | **Version details** – default tab (usually Overview). |
| `toARVersionDetailsTab({ ... params, versionTab })` | `.../versions/:versionIdentifier/:versionTab` | **Version details** – specific tab (overview, artifact_details, supply_chain, etc.). |
| `toARRepositoryWebhookDetails(...)` | `/registries/:repo/webhooks/:webhookIdentifier` | **Webhook details**. |
| `toARRepositoryWebhookDetailsTab(...)` | `.../webhooks/:id/:tab` | **Webhook details** – tab. |

**Note**: In **Gitness** (OSS), route definitions may be under a different base (e.g. space); see `gitness/utils/getARRouteDefinitions.ts` and `giness/RouteDefinitions.ts`.

### 3.2 Route destinations (what renders where)

- **File**: `routes/RouteDestinations.tsx`

**Logic**:

- **`/`** → redirect to `toARRepositories()`.
- **`/redirect`** → **RedirectPage**.
- **`/registries`** (exact):
  - If **Directory** view → **RepositoryListTreeViewPage**.
  - If **List** view → **RepositoryListPage**.
- **`/registries/:repositoryIdentifier`** and **`/registries/:repositoryIdentifier/:tab`** → **RepositoryDetailsPage** (public routes).
- **Webhook routes** → **WebhookDetailsPage**.
- **Version details routes** (multiple path variants for org/project/pipeline/ssca) → **VersionDetailsPage** (only when `parent === Enterprise` or `repositoryListViewType === Directory`; i.e. “separate version details route”).
- **Artifact routes** (`.../versions`, `.../properties`, `.../[artifactType]/[artifactIdentifier]`) → **ArtifactDetailsPage** (which internally can show version list or version details in OSS when not using separate version route).

So:

- **Repository list** (list or tree) → **RepositoryListPage** or **RepositoryListTreeViewPage**.
- **Repository details** (tabs: Packages/Datasets/Models/Configuration/Webhooks/Metadata) → **RepositoryDetailsPage**.
- **Artifact details** (tabs: Versions, Metadata) → **ArtifactDetailsPage**.
- **Version details** (tabs: Overview, Artifact details, Supply chain, etc.) → **VersionDetailsPage** (when using separate route) or embedded under **ArtifactDetailsPage** (OSS list view).

---

## 4. Page Hierarchy & UI Flow

High-level flow:

```
App
 └─ NavComponent
     └─ AsyncDownloadRequestsProvider
         └─ RouteDestinations (Switch)
             ├─ Repository List (list or tree)
             ├─ Repository Details (header + tabs)
             │    ├─ Packages / Datasets / Models → artifact list for that repo (and type)
             │    ├─ Configuration → repo config form
             │    ├─ Webhooks → webhook list
             │    └─ Metadata → properties form
             ├─ Artifact Details (header + tabs)
             │    ├─ Versions → version list table (package-type-specific columns/actions)
             │    └─ Metadata → properties
             │    (When separate version route is used, version details live on VersionDetailsPage.)
             ├─ Version Details (header + tabs)
             │    ├─ Overview → package-type-specific overview (e.g. SwiftOverviewPage)
             │    ├─ Artifact details → package-type-specific artifact details (e.g. readme, files, dependencies)
             │    ├─ Supply chain / Security tests / Deployments / Code (Enterprise, feature-flagged)
             │    └─ OSS (open source view)
             ├─ Webhook Details
             └─ Redirect
```

---

## 5. Factory Pattern (Repository & Version)

The UI is **package-type-aware** via two factory patterns. Each **package type** (DOCKER, NPM, SWIFT, etc.) has:

1. A **Repository** type (e.g. `SwiftRepositoryType`) – defines repo creation form, config form, setup client, tabs, tree node, etc.
2. A **Version** type (e.g. `SwiftVersionType`) – defines version list table, version details header, version details tabs (overview, artifact details, etc.), artifact/version actions, tree views, row expand (files).

### 5.1 Version factory

- **Framework base**: `frameworks/Version/Version.tsx` – abstract class **VersionStep\<T\>**.
  - Key abstract props: `packageType`, `allowedVersionDetailsTabs`, `hasArtifactRowSubComponent`.
  - Abstract render methods: `renderVersionListTable`, `renderVersionDetailsHeader`, `renderVersionDetailsTab`, `renderArtifactActions`, `renderVersionActions`, `renderArtifactRowSubComponent`, `renderArtifactTreeNodeView`, `renderArtifactTreeNodeDetails`, `renderVersionTreeNodeView`, `renderVersionTreeNodeDetails`.
- **Framework factory**: `frameworks/Version/VersionAbstractFactory.tsx` (base), `frameworks/Version/VersionFactory.tsx` (singleton).
- **Registration**: `pages/version-details/VersionFactory.tsx` – imports **versionFactory** and registers every *VersionType* (Docker, Helm, Generic, Maven, Npm, Python, NuGet, RPM, Cargo, Go, Huggingface, Conda, Dart, Composer, **Swift**).

**Widgets** (resolve package type and delegate to the registered step):

- **VersionListTableWidget** – `versionFactory.getVersionType(packageType).renderVersionListTable(...)` → **Version list table** on Artifact Details → Versions tab.
- **VersionDetailsHeaderWidget** – renders **Version details header** (title, breadcrumbs).
- **VersionDetailsTabWidget** – renders **Version details tab content** (Overview, Artifact details, etc.) by `versionTab` from URL.

So: **Version list**, **Version details header**, and **Version details tab content** are all chosen by `packageType` from the factory.

### 5.2 Repository factory

- **Framework base**: `frameworks/RepositoryStep/Repository.tsx` – abstract class **RepositoryStep\<T, U\>**.
  - Key abstract props: `packageType`, `repositoryName`, `defaultValues`, `defaultUpstreamProxyValues`, `repositoryIcon`, `supportsUpstreamProxy`, optional `supportedScanners`, `isWebhookSupported`, `supportedRepositoryTabs`, etc.
  - Abstract render methods: `renderCreateForm`, `renderCofigurationForm`, `renderActions`, `renderSetupClient`, `renderRepositoryDetailsHeader`, `renderRedirectPage`, `renderTreeNodeView`, `renderTreeNodeDetails`.
- **Framework factory**: `frameworks/RepositoryStep/RepositoryAbstractFactory.tsx`, `frameworks/RepositoryStep/RepositoryFactory.tsx` (singleton).
- **Registration**: `pages/repository-details/RepositoryFactory.tsx` – registers every *RepositoryType* (Docker, Maven, Helm, … Swift).

**Usage**:

- **Repository details tabs**: `repositoryFactory.getRepositoryType(data?.packageType)` → `getSupportedRepositoryTabs()` → filters **RepositoryDetailsTabs** (Packages, Configuration, Webhooks, Metadata, etc.).
- **Repository configuration form**: **RepositoryConfigurationFormWidget** uses repository type to render the correct config form.
- **Create repository**: Package type selector and create flow use repository types (e.g. from `hooks/useGetRepositoryTypes.ts`).

---

## 6. Key Pages in Detail (with UI)

### 6.1 Repository list

- **List view**: `pages/repository-list/RepositoryListPage.tsx`
  - **UI**: Page header (title, breadcrumbs), filters (search, package type, repository type, config type, scope, soft delete), list/table toggle, **RepositoryListTable**, “Create repository” button.
  - **Data**: `useLocalGetRegistriesQuery` (or equivalent) with query params (search, page, size, repositoryTypes, configType, scope, softDeleteFilter).
- **Tree view**: `pages/repository-list/RepositoryListTreeViewPage.tsx`
  - **UI**: Header, package type / config type filters, **RepositoryListTreeView** (tree of repos → artifacts → versions). Clicking a node navigates (e.g. `toARRepositoryDetailsTab`, `toARArtifactDetails`, `toARVersionDetailsTab`).

### 6.2 Repository details

- **Page**: `pages/repository-details/RepositoryDetailsPage.tsx`
  - Structure: **RepositoryProvider** → **RepositoryHeader** + **RepositoryDetails**.
- **RepositoryProvider** (`context/RepositoryProvider.tsx`): Fetches registry by `repositoryIdentifier` (and scope), provides `data`, `isReadonly`, `refetch`, `setIsDirty`, `setIsUpdating` to children.
- **RepositoryDetails** (`RepositoryDetails.tsx`): Renders **tabs** (Packages, Datasets, Models, Configuration, Webhooks, Metadata) based on `repositoryFactory.getRepositoryType(data?.packageType).getSupportedRepositoryTabs()` and feature flags / parent. Active tab drives route (`repositoryDetailsTabPathProps`). Tab content is rendered by **RepositoryDetailsTabPage**.
- **RepositoryDetailsTabPage** (`RepositoryDetailsTabPage.tsx`):
  - **PACKAGES** → **RegistryArtifactListPage** (artifact list for this repo, `artifactType=ARTIFACTS`).
  - **DATASETS** / **MODELS** → same with `artifactType=DATASET` / `MODEL`.
  - **CONFIGURATION** → **RepositoryConfigurationFormWidget** (package-type-specific form).
  - **WEBHOOKS** → **WebhookListPage**.
  - **METADATA** → **PropertiesFormContent**.

**UI**: Repository name/title in header, breadcrumbs, tab bar. Content = one of the above.

### 6.3 Artifact details

- **Page**: `pages/artifact-details/ArtifactDetailsPage.tsx`
  - **ArtifactProvider** (repoKey, artifact from path) → **ArtifactDetailsHeader** + **ArtifactDetails**.
- **ArtifactDetails** (`ArtifactDetails.tsx`): Tabs come from `versionFactory.getVersionType(data?.packageType).getSupportedArtifactTabs()` (typically **Versions**, **Metadata**). Tab change updates route (`toARArtifactVersions` / `toARArtifactProperties`). Content:
  - **VERSIONS** → **VersionListPage** (which uses **VersionListTableWidget** → package-type-specific **VersionListTable** with columns and actions).
  - **METADATA** → **PropertiesFormContent**.
  - When **separate version details route** is not used (e.g. OSS list view), version detail can be shown via **OSSVersionDetailsPage** or inline.

**UI**: Artifact name in header, breadcrumbs, Versions / Metadata tabs. Versions tab = table of versions (name, size, file count, download count, pull command, last modified, actions). Row expand (if `hasArtifactRowSubComponent`) shows e.g. **ArtifactFilesContent** (minimal file list).

### 6.4 Version details

- **Page**: `pages/version-details/VersionDetailsPage.tsx`
  - **VersionProvider** (repoKey, artifactKey, versionKey from path) → **VersionDetails**.
- **VersionProvider** (`context/VersionProvider.tsx`): Fetches **artifact version summary** by repo/artifact/version (and optional digest). Provides `data`, `isReadonly`, `refetch`, etc., to children.
- **VersionDetails** (`VersionDetails.tsx`): **VersionDetailsHeader** + **VersionDetailsTabs**.
  - **VersionDetailsHeader**: Uses **VersionDetailsHeaderWidget** with `data.packageType` and `data` → package-type-specific header content (e.g. **VersionDetailsHeaderContent** for Swift/Generic).
  - **VersionDetailsTabs**: Reads `versionTab` from URL; builds tab list from **VersionDetailsTabList** filtered by `versionType.getAllowedVersionDetailsTab()` and feature flags / parent. Each tab navigates to `toARVersionDetailsTab(..., versionTab)`. Tab **content** is rendered by **VersionDetailsTabWidget** → `versionFactory.getVersionType(packageType).renderVersionDetailsTab({ tab })`.

**Tab content examples (by package type)**:

- **OVERVIEW**: e.g. **SwiftOverviewPage** → **SwiftVersionGeneralInfo** (card with name, version, package type, size, downloads, uploaded by, description). Other types: **PythonVersionOverviewPage**, **NuGetVersionOverviewPage**, **GoVersionOverviewPage**, etc.
- **ARTIFACT_DETAILS**: e.g. **SwiftArtifactDetailsPage** → tabs Readme / Files / Dependencies (**SwiftVersionFilesContent**, **SwiftVersionDependencyContent**, **ReadmeFileContent**). Other types have their own artifact-details pages (NuGet, Npm, Python, etc.).
- **CODE** / **OSS**: Some types render **OSSContentPage** or similar.
- **SUPPLY_CHAIN**, **SECURITY_TESTS**, **DEPLOYMENTS**: Enterprise, feature-flagged.

**UI**: Version identifier in header, breadcrumbs (Repository → Artifact). Tab bar (Overview, Artifact details, …). Content = one of the above.

### 6.5 Global artifact list

- **Page**: `pages/artifact-list/ArtifactListPage.tsx`
  - **UI**: Header, filters (search, repository, package types, deployed artifacts, soft delete, metadata filter), **ArtifactListTable**. Used when navigating to “Artifacts” (e.g. `toARArtifacts()`).

---

## 7. Contexts & Providers

| Context / Provider | File | Purpose |
|--------------------|------|---------|
| **AppStoreContext** | `contexts/AppStoreContext.tsx` | Scope (account/org/project/space), baseUrl, matchPath, parent (OSS/Enterprise), repositoryListViewType, setRepositoryListViewType, isPublicAccessEnabledOnResources, isCurrentSessionPublic. |
| **ParentProvider** | `contexts/ParentProvider.tsx` | Injects parent’s hooks (useQueryParams, useUpdateQueryParams, usePermission, useConfirmationDialog, useModalHook, …) and components (RbacButton, ModalProvider, PageNotPublic, …). |
| **RepositoryProviderContext** | `pages/repository-details/context/RepositoryProvider.tsx` | Current repository `data`, isDirty, isUpdating, isReadonly, refetch, setIsDirty, setIsUpdating. |
| **ArtifactProviderContext** | `pages/artifact-details/context/ArtifactProvider.tsx` | Current artifact `data`, isDirty, isUpdating, isReadonly, refetch, setters. |
| **VersionProviderContext** | `pages/version-details/context/VersionProvider.tsx` | Current version summary `data`, isReadonly, refetch, isDirty, isUpdating, setters. |
| **VersionOverviewContext** | `pages/version-details/context/VersionOverviewProvider.tsx` | Used **inside** version details tab content; provides full **artifact details** `data` and refetch (e.g. for Overview and Artifact details tabs that need readme, dependencies, metadata). |
| **VersionFilesContext** | `pages/version-details/context/VersionFilesProvider.tsx` | Paginated/sortable **file list** for a version; used by “Files” tab and expandable row file list. |
| **AsyncDownloadRequestsProvider** | `contexts/AsyncDownloadRequestsProvider/AsyncDownloadRequestsProvider.tsx` | Tracks async download requests (e.g. for bulk download). |

---

## 8. Hooks (main ones)

- **File**: `hooks/index.ts` (re-exports).

| Hook | Purpose |
|------|---------|
| **useAppStore** | Access AppStoreContext (scope, parent, repositoryListViewType, etc.). |
| **useRoutes** | Get route builder (toARRepositories, toARVersionDetailsTab, …). |
| **useDecodedParams** | Decode path params (e.g. repositoryIdentifier, artifactIdentifier, versionIdentifier, tab). |
| **useParentHooks** | Parent’s useQueryParams, useUpdateQueryParams, usePreferenceStore, usePermission. |
| **useParentComponents** | Parent’s RbacButton, ModalProvider, PageNotPublic, etc. |
| **useGetSpaceRef** | Space/scope ref for API calls (account/org/project/space). |
| **useGetRepositoryListViewType** | List vs Directory view. |
| **useFeatureFlags** / **useFeatureFlag** | Feature flags (e.g. HAR_CUSTOM_METADATA_ENABLED, HAR_TRIGGERS). |
| **useAllowSoftDelete** | Whether soft-delete filter/actions are allowed. |
| **useGetRepositoryTypes** | List of repository types for selectors (from `hooks/useGetRepositoryTypes.ts` – includes **RepositoryTypes** array with label, value, icon; Swift is here with icon `maven-repository-type` until a Swift icon is added). |

---

## 9. Frameworks (Version & Repository steps)

### 9.1 Version framework (`frameworks/Version/`)

- **Version.tsx** – Abstract **VersionStep\<T\>** (see section 5.1).
- **VersionAbstractFactory.tsx** – Map of packageType → VersionStep; `registerStep`, `getVersionType`.
- **VersionFactory.tsx** – Singleton factory instance.
- **VersionListTableWidget.tsx** – Resolves version type, renders version list table.
- **VersionDetailsHeaderWidget.tsx** – Resolves version type, renders version details header.
- **VersionDetailsTabWidget.tsx** – Resolves version type, renders version details tab panel.
- **VersionActionsWidget.tsx** – Version actions (e.g. delete, setup client, quarantine, download).
- **ArtifactActionsWidget.tsx**, **ArtifactRowSubComponentWidget.tsx**, **ArtifactTreeNodeViewWidget.tsx**, **ArtifactTreeNodeDetailsWidget.tsx**, **VersionTreeNodeViewWidget.tsx**, **VersionTreeNodeDetailsWidget.tsx** – Delegate to version type for artifact/version tree and row expand.

### 9.2 Repository framework (`frameworks/RepositoryStep/`)

- **Repository.tsx** – Abstract **RepositoryStep\<T, U\>** (see section 5.2).
- **RepositoryAbstractFactory.tsx** – Map of packageType → RepositoryStep.
- **RepositoryFactory.tsx** – Singleton.
- **RepositoryConfigurationFormWidget.tsx** – Renders package-type-specific configuration form.
- **RepositoryDetailsHeaderWidget.tsx**, **CreateRepositoryWidget.tsx**, **RepositoryActionsWidget.tsx**, **RepositoryTreeNodeViewWidget.tsx**, **RepositoryTreeNodeDetailsWidget.tsx**, **RepositorySetupClientWidget.tsx**, **RedirectWidget.tsx** – All delegate to repository type.

---

## 10. Shared Components (`components/`)

Used across pages. Examples:

- **Breadcrumbs** – Breadcrumb links (e.g. Repositories > Artifacts > Version).
- **ButtonTabs** / **ButtonTab** – Tab UI used in version artifact details (Readme, Files, Dependencies).
- **TabsContainer** – Wraps tab content.
- **RouteProvider** – Wraps a Route; handles public vs non-public (PageNotPublic) and **ParentSyncProvider**.
- **PackageTypeSelector** – Filter by package type (Docker, NPM, Swift, …).
- **MetadataFilterSelector** – Custom metadata filter (e.g. HAR_CUSTOM_METADATA_ENABLED).
- **PropertiesForm** / **PropertiesFormContent** – Metadata/properties form (repository metadata, artifact properties).
- **TreeView** – Tree (e.g. repository tree view).
- **SetupClientButton** – “Setup client” for a package type.
- **CommandBlock** – Copyable command (e.g. pull command).
- **TableCells** – Reusable table cell components.
- **MultiTagsInput**, **NameDescriptionTags**, **LabelsPopover**, **ManageMetadata** – Forms and metadata UI.

---

## 11. Common Types & Constants

- **common/types.ts**: **RepositoryPackageType** (DOCKER, NPM, SWIFT, …), **RepositoryConfigType** (VIRTUAL, UPSTREAM), **Parent** (OSS, Enterprise), **PageType**, **EntityScope**, **RepositoryScopeType**, **Scanners**, etc.
- **constants** (e.g. `constants/index.ts`): **DEFAULT_PAGE_INDEX**, **PreferenceScope**, **SoftDeleteFilterEnum**, etc.
- **common/permissionTypes**: **PermissionIdentifier**, **ResourceType** (e.g. ARTIFACT_REGISTRY).

---

## 12. Strings & i18n

- **strings/strings.en.yaml** – All user-facing strings (keys like `repositoryList.pageHeading`, `packageTypes.swiftPackage`, `repositoryTypes.swift`, `versionDetails.tabs.overview`, etc.).
- **strings/types.ts** – **StringsMap** (TypeScript keys for YAML).
- **frameworks/strings/** – **StringsContextProvider**, **String** component, **useStrings** (getString), **languageLoader**.

Adding a new package type (e.g. Swift) requires: **packageTypes.swiftPackage**, **repositoryTypes.swift**, and any version/artifact tab or label strings.

---

## 13. Package Types & Adding a New One (e.g. Swift)

To add a package type (Swift is already added; this is the checklist):

1. **common/types.ts** – Add enum value e.g. `SWIFT = 'SWIFT'` to **RepositoryPackageType**.
2. **strings** – Add `packageTypes.swiftPackage`, `repositoryTypes.swift` and ensure **StringsMap** in `strings/types.ts` includes them.
3. **useGetRepositoryTypes.ts** – Add entry to **RepositoryTypes** array (label, value, icon). (Swift currently uses `maven-repository-type`; see `SWIFT_ICON_SETUP.md` for a dedicated icon.)
4. **Version type** – Under `pages/version-details/SwiftVersion/`:
   - **SwiftVersionType.tsx** – Extend **VersionStep**, implement all render* methods, set `allowedVersionDetailsTabs` (e.g. OVERVIEW, ARTIFACT_DETAILS), `versionListTableColumnConfig`, `allowedActionsOnVersion` / `allowedActionsOnVersionDetailsPage`.
   - **types.ts** – e.g. **SwiftArtifactDetailsTabEnum**, **SwiftVersionDetailsQueryParams**, **SwiftArtifactDetails** (extends ArtifactDetail with local metadata).
   - **pages/overview/** – **SwiftOverviewPage**, **SwiftVersionGeneralInfo** (overview card).
   - **pages/artifact-details/** – **SwiftArtifactDetailsPage** (tabs: Readme, Files, Dependencies), **SwiftVersionFilesContent**, **SwiftVersionDependencyContent**.
5. **VersionFactory.tsx** – `versionFactory.registerStep(new SwiftVersionType())`.
6. **Repository type** – Under `pages/repository-details/SwiftRepository/`:
   - **SwiftRepositoryType.tsx** – Extend **RepositoryStep**, set defaultValues, defaultUpstreamProxyValues, implement render* (create form, config form, setup client, header, tree node, etc.).
7. **RepositoryFactory.tsx** – `repositoryFactory.registerStep(new SwiftRepositoryType())`.

Result in UI: Swift appears in package type filter and repository type selector; creating a Swift repo uses Swift repository form; repository details show Swift-specific tabs; artifact versions table and version details (overview, readme, files, dependencies) are Swift-specific.

---

## 14. Directory Structure Summary

```
web/src/ar/
├── app/                    # App bootstrap, OpenAPI client
├── routes/                 # Route definitions, path params, RouteDestinations
├── contexts/               # AppStoreContext, AsyncDownloadRequestsProvider
├── hooks/                  # useRoutes, useDecodedParams, useAppStore, useGetRepositoryTypes, ...
├── common/                 # types (RepositoryPackageType, etc.), utils, permissionTypes
├── constants/              # DEFAULT_PAGE_INDEX, PreferenceScope, SoftDeleteFilterEnum
├── frameworks/
│   ├── Version/            # VersionStep, VersionFactory, *Widget (list, header, tab, actions, tree)
│   ├── RepositoryStep/    # RepositoryStep, RepositoryFactory, *Widget
│   └── strings/            # StringsContextProvider, String, useStrings
├── pages/
│   ├── repository-list/    # RepositoryListPage, RepositoryListTreeViewPage, table/tree components
│   ├── repository-details/ # RepositoryDetailsPage, RepositoryDetails, RepositoryProvider, tab page, *RepositoryType
│   ├── artifact-list/      # ArtifactListPage (global artifacts)
│   ├── artifact-details/   # ArtifactDetailsPage, ArtifactDetails, ArtifactProvider, constants
│   ├── version-list/       # VersionListPage (versions for one artifact)
│   ├── version-details/    # VersionDetailsPage, VersionDetails, VersionProvider, VersionOverviewProvider,
│   │   │                   # VersionFilesProvider, VersionDetailsTabs, VersionDetailsHeader,
│   │   │                   # *VersionType (Docker, Npm, Swift, ...), shared version-details components
│   ├── webhook-details/    # WebhookDetailsPage
│   ├── upstream-proxy-details/
│   ├── manage-registries/
│   └── redirect-page/
├── components/             # Breadcrumbs, ButtonTabs, RouteProvider, PackageTypeSelector, TreeView, ...
├── strings/                # strings.en.yaml, types (StringsMap)
├── utils/                  # queryClient, customYupValidators, ...
├── __mocks__/              # PageNotPublic, DefaultNavComponent, RbacButton, hooks (useQueryParams, ...)
├── MFEAppTypes.ts          # Parent props, Scope, AppstoreContext, ParentContextObj, components/hooks types
└── gitness/                # Gitness-specific route definitions and helpers
```

---

## 15. Data Flow (short)

- **Repository list**: Query params (search, type, scope, …) → `useLocalGetRegistriesQuery` → **RepositoryListTable** or tree.
- **Repository details**: `repositoryIdentifier` from URL → **RepositoryProvider** → `useGetRegistryQuery` → context `data` → **RepositoryDetails** tabs → **RepositoryDetailsTabPage** (artifact list / config form / webhooks / metadata).
- **Artifact details**: `repositoryIdentifier`, `artifactIdentifier`, `artifactType` → **ArtifactProvider** → artifact summary → **VersionListPage** → **VersionListTableWidget** → package-type **VersionListTable** (versions query).
- **Version details**: Path params → **VersionProvider** (`useGetArtifactVersionSummaryQuery`) → **VersionOverviewProvider** (`useGetArtifactDetailsQuery` for full artifact details) → **VersionDetailsHeader** + **VersionDetailsTabs** → **VersionDetailsTabWidget** → package-type `renderVersionDetailsTab` → e.g. **SwiftOverviewPage** + **SwiftArtifactDetailsPage** (with **VersionFilesProvider** for files).

---

This document covers the **entire structure** of the AR repo with **reference to the UI**: entry point, routing, every major page, the factory pattern, contexts, hooks, frameworks, components, types, strings, and how adding a package type (e.g. Swift) fits in. Use it as your single reference for navigation and implementation.
