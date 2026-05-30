# Homepage audit

## Scope

Audit of the current OmePic public homepage (`UploadPageClient`) before the style-repair implementation pass.

## Findings

1. The homepage currently has too many equally weighted cards near the fold.
   - The hero card, storage selector card, dropzone, token card, status card, and latest-result card all compete for attention.
   - The upload action should dominate the page, but the current structure spreads emphasis too evenly.

2. The layout is technically consistent with the broader design system, but the composition is not.
   - Shared glassmorphism, gradients, and panel styles are already present.
   - The issue is mainly hierarchy, spacing rhythm, and grouping rather than a missing theme.

3. The homepage already has a strong local reference to follow.
   - `img.html` shows the accepted visual direction: focused upload zone, restrained supporting actions, soft background glow, and clearer primary/secondary grouping.

4. There are directly related shared style defects.
   - `frontend/src/components/ui/Button.tsx` uses `px-4.5`, which is not a default Tailwind spacing token.
   - `frontend/src/components/shared/AppHeader.tsx` uses `h-4.5` / `w-4.5`, which are also invalid default Tailwind spacing tokens.
   - These invalid utilities can cause partial visual regressions and should be normalized during the pass.

## Implementation thesis

Treat this as a medium homepage restructure, not a full redesign:

* Keep the existing glass/slate/violet-cyan system.
* Make the upload workspace the obvious focal area.
* Compress or regroup secondary operational panels.
* Improve scanning order on desktop and stacking order on mobile.
* Fix invalid shared utility classes encountered in the homepage shell.

## Verification focus

* Header still reads clearly after shared fixes.
* Homepage remains fully functional for storage selection, upload initiation, progress, success result, and recent upload preview.
* Chinese and English copy still fit the revised layout without overflow pressure.
