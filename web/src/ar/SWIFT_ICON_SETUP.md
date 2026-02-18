# Adding Swift logo (SVG) so it appears everywhere in the frontend

Icons in the AR (Artifact Registry) frontend come from **`@harnessio/icons`**. To have the Swift logo show everywhere (repository type selector, repository/version trees, headers, overview cards, etc.), use one of the two approaches below.

---

## Option A: Add the icon to the Harness Icons package (recommended)

This makes Swift use the same `<Icon name="swift-icon" />` pattern as Docker, Maven, NuGet, etc., so no AR code changes are needed beyond the icon name.

### 1. Get the Swift SVG

- Use your Swift logo SVG file (e.g. `swift-logo.svg`).
- Prefer a 24×24 (or square) viewBox so it scales well.
- Keep the file small and valid (no scripts, clean paths).

### 2. Add it to the Harness Icons package

The `@harnessio/icons` package is **external** (npm). Its types say:

- *"This file is auto-generated. Please do not modify this file manually."*
- *"Use the command `yarn ui:icons` to regenerate this file."*

So you need to add the Swift icon in the **repo that owns the Harness icons** (e.g. a design-system or ui-icons repo), not in gitness.

Typical steps there:

1. Clone the Harness icons (or design-system) repo.
2. Add your SVG (e.g. `swift-icon.svg`) in the folder where other icon SVGs live (e.g. `src/icons/` or similar).
3. Register the new icon in the config/list that the icon generator reads (if applicable).
4. Run the icon generation command (e.g. `yarn ui:icons` or the script that generates `HarnessIcons` and `HarnessIconName`).
5. Ensure the new icon gets a **kebab-case name** (e.g. `swift-icon`).
6. Build and publish a new version of `@harnessio/icons`.

### 3. Use the new icon in gitness (AR)

1. **Bump the dependency** in `web/package.json`:
   - Set `"@harnessio/icons": "^x.y.z"` to the new version that includes the Swift icon.

2. **Install:**
   ```bash
   cd web && yarn install
   ```

3. **Use the new icon name everywhere Swift is referenced:**
   - **Repository type list:**  
     `web/src/ar/hooks/useGetRepositoryTypes.ts`  
     Set the Swift entry’s `icon` to the new name (e.g. `'swift-icon'`).
   - **Swift repository type:**  
     `web/src/ar/pages/repository-details/SwiftRepository/SwiftRepositoryType.tsx`  
     Set `protected repositoryIcon: IconName = 'swift-icon'`.
   - **Swift version type:**  
     `web/src/ar/pages/version-details/SwiftVersion/SwiftVersionType.tsx`  
     Any place that passes an icon (e.g. for tree or header), use `'swift-icon'`.
   - **Overview / general info:**  
     `web/src/ar/pages/version-details/SwiftVersion/pages/overview/SwiftVersionGeneralInfo.tsx`  
     In `LabelValueContent` for package type, set `icon="swift-icon"`.

After that, the Swift logo will appear everywhere the app uses `RepositoryIcon` or the same icon name (repository selector, trees, headers, overview, etc.).

---

## Option B: Use a local SVG in gitness only (no change to `@harnessio/icons`)

If you cannot change the Harness Icons package, you can host the Swift SVG in gitness and render it only where you control the component.

### 1. Add the SVG in the repo

- Example: `web/src/ar/assets/swift-logo.svg`  
  (or `web/src/icons/swift-logo.svg` if you prefer to keep it with other app icons.)

### 2. Declare SVG module (if needed)

- The AR app already has in `web/src/ar/global.d.ts`:
  ```ts
  declare module '*.svg' {
    const value: string
    export default value
  }
  ```
- If you import the SVG as URL (e.g. `import svg from './swift-logo.svg?url'`), ensure your bundler supports `?url` for SVGs (see e.g. `TreeNode.tsx`).

### 3. Create a small Swift logo component

Example: `web/src/ar/components/SwiftIcon/SwiftIcon.tsx`

```tsx
import React from 'react'
import { IconProps } from '@harnessio/icons'

import swiftLogoSvg from '@ar/assets/swift-logo.svg'  // or path you chose

interface SwiftIconProps {
  size?: number
  className?: string
}

export default function SwiftIcon({ size = 24, className }: SwiftIconProps): JSX.Element {
  return (
    <img
      src={swiftLogoSvg}
      alt="Swift"
      width={size}
      height={size}
      className={className}
      style={{ display: 'block' }}
    />
  )
}
```

(If your setup uses `?url` for SVGs, use `import swiftLogoSvg from '...svg?url'` and same `src={swiftLogoSvg}`.)

### 4. Use it where you control the UI

- **RepositoryIcon**  
  In `web/src/ar/frameworks/RepositoryStep/RepositoryIcon.tsx`, when `packageType === RepositoryPackageType.SWIFT`, return `<SwiftIcon size={...} />` instead of `<Icon name={...} />`. For all other types, keep using `<Icon name={repositoryType.getIconName()} ... />`.

- **SwiftRepositoryType**  
  `repositoryIcon` is typed as `IconName`, so you can’t pass a component there. You can keep using a placeholder `IconName` (e.g. `'generic-repository-type'`) for the type, and rely on `RepositoryIcon` to show the real Swift logo when it sees `SWIFT`.

- **Create-repository list**  
  The list uses `ButtonOption` with `icon: IconName`. So until the icon exists in `@harnessio/icons`, the create-repository grid will keep showing whatever icon name you put in `useGetRepositoryTypes` (e.g. a generic icon). Only after adding the icon to the package (Option A) will that grid show the Swift logo.

- **Version/artifact trees and headers**  
  Where the code uses `RepositoryIcon` or a single place that resolves the icon by package type, the Swift logo will appear once `RepositoryIcon` is updated as above. Any place that uses `repositoryType.getIconName()` and then `<Icon name={...} />` will still show the placeholder until the icon is in the package.

So with Option B you get the Swift logo in:
- RepositoryIcon (headers, trees, any usage of that component)
- Any other spot where you explicitly render `<SwiftIcon />` instead of `<Icon name="..." />`.

You do **not** get the Swift logo in the create-repository selector grid unless you add the icon to the Harness Icons package (Option A).

---

## Summary

| Goal                         | Approach                    |
|-----------------------------|-----------------------------|
| Swift logo everywhere       | Add SVG to Harness Icons (Option A), then use `swift-icon` in AR. |
| Swift logo in AR only       | Add local SVG + `SwiftIcon` and use it in `RepositoryIcon` and any custom screens (Option B). |
| Create-repository grid icon  | Requires the icon in `@harnessio/icons` (Option A). |

Recommended: add the Swift SVG to the Harness Icons package, publish a new version, then in gitness bump `@harnessio/icons` and set all Swift-related `icon` / `repositoryIcon` to the new name (e.g. `'swift-icon'`).
