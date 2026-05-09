# Research: shadcn/ui Ecosystem for Image Hosting Frontend Rebuild

- **Query**: Comprehensive research on shadcn/ui ecosystem -- official components, third-party libraries, image-hosting-specific components, and theming best practices
- **Scope**: External (web research via proxy) + Internal (existing project code analysis)
- **Date**: 2026-05-06

---

## 1. Official shadcn/ui v4 Registry -- All Available Components

The project currently uses shadcn/ui v2-era components (manually written, not from the CLI). The official shadcn/ui has since moved to v4 with a completely revamped registry at `https://ui.shadcn.com/r/styles/default/{name}.json`. The v4 registry contains **54 UI components**, **97 blocks**, **5 themes**, and **1 hook**.

### 1.1 Components We Already Have (12)

| Component | Current File | Notes |
|---|---|---|
| Button | `frontend/src/components/ui/Button.tsx` | Custom variant additions (primary, danger) |
| Card | `frontend/src/components/ui/Card.tsx` | Custom variant system (default, strong, subtle) |
| Dialog | `frontend/src/components/ui/Dialog.tsx` | Custom closeLabel, showCloseButton props |
| DropdownMenu | `frontend/src/components/ui/DropdownMenu.tsx` | Standard shadcn implementation |
| Input | `frontend/src/components/ui/Input.tsx` | Custom hover/focus ring styling |
| Label | `frontend/src/components/ui/Label.tsx` | Standard shadcn implementation |
| Select | `frontend/src/components/ui/Select.tsx` | Standard shadcn implementation |
| Separator | `frontend/src/components/ui/Separator.tsx` | Standard shadcn implementation |
| Table | `frontend/src/components/ui/Table.tsx` | Standard shadcn implementation |
| Tabs | `frontend/src/components/ui/Tabs.tsx` | Standard shadcn implementation |
| Badge | `frontend/src/components/ui/Badge.tsx` | Standard shadcn implementation |
| Textarea | `frontend/src/components/ui/Textarea.tsx` | Custom hover/focus ring styling |

### 1.2 Official Components NOT Yet in the Project (42)

These are available from the official shadcn/ui v4 registry and can be added via `npx shadcn@latest add <name>`.

#### High-Priority for Image Hosting App

| Component | Radix Dependency | Description |
|---|---|---|
| **skeleton** | none | `animate-pulse rounded-md bg-accent` -- loading placeholders for image cards, tables |
| **progress** | `@radix-ui/react-progress` | Progress bar -- upload progress visualization |
| **tooltip** | `@radix-ui/react-tooltip` | Hover tooltips -- copy button labels, image info |
| **popover** | `@radix-ui/react-popover` | Rich popovers -- image detail previews, format selector |
| **scroll-area** | `@radix-ui/react-scroll-area` | Custom scrollbars -- admin tables, history lists |
| **avatar** | `@radix-ui/react-avatar` | User avatars with fallback -- admin user display |
| **empty** | none | Empty state container -- "no images uploaded yet" |
| **pagination** | none | Page navigation -- admin image list pagination |
| **sheet** | `@radix-ui/react-dialog` | Slide-out panel -- mobile sidebar, image detail drawer |
| **sonner** | `sonner`, `next-themes` | Toast notifications -- shadcn's recommended toast (vs react-hot-toast) |
| **spinner** | `class-variance-authority` | Loading spinner -- upload in progress, data fetching |

#### Medium-Priority

| Component | Radix Dependency | Description |
|---|---|---|
| **drawer** | `vaul`, `@radix-ui/react-dialog` | Bottom drawer -- mobile image actions |
| **command** | `cmdk` | Command palette -- keyboard shortcuts, quick actions |
| **collapsible** | `@radix-ui/react-collapsible` | Expandable sections -- admin settings groups |
| **hover-card** | `@radix-ui/react-hover-card` | Rich hover preview -- image metadata on hover |
| **checkbox** | `@radix-ui/react-checkbox` | Checkbox -- multi-select images for batch delete |
| **radio-group** | `@radix-ui/react-radio-group` | Radio buttons -- storage backend selection |
| **switch** | `@radix-ui/react-switch` | Toggle switch -- dark mode, settings toggles |
| **toggle** | `@radix-ui/react-toggle` | Toggle button -- view mode toggle (grid/list) |
| **toggle-group** | `@radix-ui/react-toggle-group` | Toggle group -- format filter group |
| **breadcrumb** | `@radix-ui/react-slot` | Breadcrumb nav -- admin navigation |
| **accordion** | `@radix-ui/react-accordion` | Accordion -- FAQ, help sections |
| **alert** | none | Alert banner -- error messages, warnings |
| **alert-dialog** | `@radix-ui/react-alert-dialog` | Confirmation dialog -- delete image confirmation |
| **aspect-ratio** | `@radix-ui/react-aspect-ratio` | Fixed aspect ratio container -- image thumbnails |
| **kbd** | none | Keyboard shortcut display -- shortcut hints |
| **item** | none | List item primitive |
| **field** | none | Form field wrapper |
| **input-group** | none | Input with addons -- URL input with copy button |
| **button-group** | none | Button groups -- format selector group |

#### Lower-Priority (Admin/Dashboard Focused)

| Component | Radix Dependency | Description |
|---|---|---|
| **sidebar** | `@radix-ui/react-slot`, `class-variance-authority`, `lucide-react` | Full sidebar system -- admin dashboard layout |
| **form** | `@radix-ui/react-label`, `@radix-ui/react-slot`, `@hookform/resolvers`, `zod`, `react-hook-form` | Form validation -- admin settings forms |
| **calendar** | `react-day-picker@latest`, `date-fns` | Date picker -- date range filter for history |
| **chart** | `recharts@2.15.4`, `lucide-react` | Charts -- admin dashboard analytics |
| **carousel** | `embla-carousel-react` | Image carousel -- featured images |
| **resizable** | `react-resizable-panels` | Resizable panels -- admin layout |
| **input-otp** | `input-otp` | OTP input -- admin 2FA |
| **menubar** | `@radix-ui/react-menubar` | Menu bar -- desktop app-like menus |
| **navigation-menu** | `@radix-ui/react-navigation-menu` | Nav menu -- site navigation |
| **context-menu** | `@radix-ui/react-context-menu` | Right-click menu -- image context actions |
| **slider** | `@radix-ui/react-slider` | Range slider -- quality/compression settings |
| **native-select** | none | Native select -- simple dropdown alternative |

### 1.3 Official Blocks (97 total)

Key blocks relevant to an image hosting app:

- **dashboard-01**: A dashboard with sidebar, charts and data table -- useful for admin panel
- **sidebar-01 through sidebar-16**: 16 sidebar variants (simple, collapsible, floating, with submenus, file tree, calendar, etc.)
- **login-01 through login-05**: 5 login page variants
- **signup-01 through signup-05**: 5 signup page variants
- **chart-*** blocks: 60+ chart variants (area, bar, line, pie, radar, radial, tooltip)

### 1.4 Official Themes (5)

All use HSL color space with CSS variables. Available themes: `theme-stone`, `theme-zinc`, `theme-neutral`, `theme-gray`, `theme-slate`. Each provides light + dark mode via `.dark .theme-{name}` selector.

---

## 2. Third-Party shadcn/ui Component Libraries

The official shadcn/ui directory.json lists **170+ third-party registries**. Below are the most notable ones.

### 2.1 Aceternity UI (ui.aceternity.com)

- **106 components**, 9 blocks, 1 hook, 165 examples
- **Install**: `npx shadcn@latest add @aceternity/{component}`
- **Tech stack**: React 18+, Next.js 13+, Tailwind CSS v4, Framer Motion, TypeScript
- **Last updated**: 2026-05-06

#### Categories and Components

**Backgrounds** (12): `background-beams`, `background-beams-with-collision`, `background-boxes`, `background-gradient`, `background-gradient-animation`, `background-lines`, `background-ripple-effect`, `aurora-background`, `wavy-background`, `dotted-glow-background`, `noise-background`, `scales`

**Text Effects** (8): `text-generate-effect`, `text-reveal-card`, `text-hover-effect`, `typewriter-effect`, `flip-words`, `colourful-text`, `encrypted-text`, `cover`

**Cards** (11): `3d-card`, `card-hover-effect`, `card-spotlight`, `card-stack`, `evervault-card`, `glare-card`, `wobble-card`, `comet-card`, `tooltip-card`, `focus-cards`, `draggable-card`

**Navigation** (6): `floating-navbar`, `navbar-menu`, `sidebar`, `floating-dock`, `resizable-navbar`, `tabs`

**Hero Sections** (7): `hero-parallax`, `hero-highlight`, `spotlight`, `spotlight-new`, `lamp`, `vortex`, `container-scroll-animation`

**Animations** (14): `animated-modal`, `animated-testimonials`, `animated-tooltip`, `apple-cards-carousel`, `carousel`, `infinite-moving-cards`, `moving-border`, `parallax-scroll`, `parallax-scroll-2`, `tracing-beam`, `following-pointer`, `layout-text-flip`, `container-text-flip`, `3d-marquee`

**Effects** (17): `sparkles`, `glowing-stars`, `meteors`, `shooting-stars`, `stars-background`, `canvas-reveal-effect`, `svg-mask-effect`, `glowing-effect`, `pointer-highlight`, `lens`, `compare`, `direction-aware-hover`, `hover-border-gradient`, `pixelated-canvas`, `dither-shader`, `webcam-pixel-grid`

**Layout** (6): `bento-grid`, `layout-grid`, `sticky-scroll-reveal`, `timeline`, `grid`, `moving-line`

**Forms** (5): `input`, `label`, `file-upload`, `placeholders-and-vanish-input`, `gooey-input`

**Utilities** (13): `3d-pin`, `macbook-scroll`, `globe`, `world-map`, `link-preview`, `multi-step-loader`, `loader`, `stateful-button`, `code-block`, `sticky-banner`, `images-slider`, `google-gemini-effect`, `tailwindcss-buttons`

**New/Recent**: `3d-globe`, `ascii-art`, `canvas-text`, `images-badge`, `keyboard`, `terminal`, `text-flipping-board`

#### Most Relevant for Image Hosting

- **file-upload**: Drag-and-drop file upload with background grid and micro-interactions (uses `react-dropzone`, `motion`)
- **lens**: Zoom into images on hover -- perfect for image thumbnails
- **compare**: Before/after image comparison slider
- **focus-cards**: Hover to focus one card, blurring others -- gallery effect
- **images-slider**: Full-page image slider with keyboard navigation
- **glare-card**: Glare effect on hover (like Linear) -- premium feel for image cards
- **card-spotlight**: Spotlight effect revealing radial gradient on hover
- **stateful-button**: Button with loading/success states -- upload button
- **multi-step-loader**: Step loader for long operations -- upload processing
- **loader**: Set of minimal loaders -- image loading placeholders
- **sparkles**: Configurable sparkle effects -- success celebration after upload
- **animated-modal**: Animated modal with transitions -- image preview alternative
- **layout-grid**: Animated grid layout -- gallery view
- **bento-grid**: Bento grid layout -- feature showcase
- **background-beams**: Animated background beams -- hero section enhancement

### 2.2 Magic UI (magicui.design)

- **74 components**, many examples
- **Install**: `npx shadcn@latest add @magicui/{component}`
- Focus on animations, micro-interactions, and visual effects

#### Key Components

**Image/Media**: `lens` (image zoom), `pixel-image` (pixelated image effect), `hero-video-dialog`, `safari` (browser mockup), `iphone-15-pro` (device mockup)

**Buttons**: `shimmer-button`, `shiny-button`, `rainbow-button`, `pulsating-button`, `ripple-button`, `interactive-hover-button`, `animated-subscribe-button`

**Text Effects**: `aurora-text`, `morphing-text`, `line-shadow-text`, `animated-shiny-text`, `animated-gradient-text`, `sparkles-text`, `spinning-text`, `flip-text`, `comic-text`, `hyper-text`, `typing-animation`, `word-rotate`, `text-reveal`, `text-animate`, `scroll-based-velocity`, `box-reveal`, `highlighter`

**Backgrounds/Patterns**: `grid-pattern`, `dot-pattern`, `striped-pattern`, `interactive-grid-pattern`, `animated-grid-pattern`, `flickering-grid`, `retro-grid`, `grid-beams`, `warp-background`

**Cards/Containers**: `magic-card` (spotlight border effect), `neon-gradient-card`, `bento-grid`, `shine-border`

**Animations/Effects**: `meteors`, `particles`, `orbiting-circles`, `border-beam`, `animated-beam`, `ripple`, `confetti`, `number-ticker`, `scroll-progress`, `blur-fade`, `progressive-blur`, `marquee`, `cool-mode`

**Utilities**: `script-copy-btn` (copy to clipboard), `dock` (macOS dock), `avatar-circles`, `icon-cloud` (3D tag cloud), `file-tree`, `terminal`, `animated-theme-toggler`, `smooth-cursor`, `pointer`, `scratch-to-reveal`, `animated-circular-progress-bar`, `animated-list`, `arc-timeline`

#### Most Relevant for Image Hosting

- **lens**: Interactive image zoom -- thumbnail hover preview
- **script-copy-btn**: Copy to clipboard -- URL/Markdown/BBcode copy
- **confetti**: Celebration effect -- after successful upload
- **number-ticker**: Animated number -- upload count, storage stats
- **animated-circular-progress-bar**: Circular progress -- upload progress
- **blur-fade**: Smooth content transitions -- image loading
- **marquee**: Infinite scroll -- image showcase banner
- **shine-border**: Animated border -- highlight featured images
- **magic-card**: Spotlight border on hover -- premium image cards
- **pixel-image**: Pixelated image effect -- loading placeholder or style effect
- **safari/iphone-15-pro**: Device mockups -- showcase images in context

### 2.3 Other Notable Third-Party Libraries

From the official directory.json (170+ registries), these are particularly relevant:

| Registry | URL | Focus |
|---|---|---|
| **@shadcnblocks** | shadcnblocks.com | Pre-built page blocks/sections |
| **@cult-ui** | cult-ui.com | Premium components and templates |
| **@eldoraui** | eldoraui.site | Animated components |
| **@kokonutui** | kokonutui.com | UI components and blocks |
| **@motion-primitives** | motion-primitives.com | Animation primitives |
| **@react-bits** | reactbits.dev | Small, focused React components |
| **@originui** | originui.com | Copy-paste UI components |
| **@shadcncraft** | shadcncraft.com | Crafted components |
| **@lmscn** | lmscn.vercel.app | Component collection |
| **@animate-ui** | animate-ui.com | Animation-focused components |
| **@bundui** | bundui.io | Component bundle |
| **@boldkit** | boldkit.dev | Bold design components |
| **@retroui** | retroui.dev | Retro-styled components |
| **@neobrutalism** | neobrutalism.dev | Neo-brutalist design |
| **@plate** | platejs.org | Rich text editor |
| **@assistant-ui** | assistant-ui.com | AI assistant components |
| **@better-upload** | better-upload.com | File upload components |
| **@evilcharts** | evilcharts.com | Chart components |
| **@lucide-animated** | lucide-animated.com | Animated Lucide icons |
| **@loading-ui** | loading-ui.com | Loading state components |
| **@spell** | spell.sh | AI-powered components |
| **@tailwind-builder** | tailwindbuilder.ai | Visual builder |

---

## 3. Image-Hosting-Specific Component Recommendations

### 3.1 File Upload with Drag-and-Drop

**Current state**: The project has a custom `UploadDropzone.tsx` with manual drag/drop handling, drag depth tracking, and a hidden file input. It works but is custom-built.

**Available options**:

1. **Aceternity `file-upload`**: Built on `react-dropzone` + `motion`. Provides drag-and-drop with background grid pattern, file preview, and micro-interactions. Dependencies: `@tabler/icons-react`, `react-dropzone`, `motion`.

2. **`@better-upload` registry**: Dedicated file upload components from better-upload.com.

3. **Keep custom + enhance**: The current `UploadDropzone` is well-built with proper accessibility (keyboard reachable, ARIA labels). Could enhance with:
   - `motion` (Framer Motion) for smoother drag state transitions
   - shadcn `progress` component for upload progress
   - shadcn `sonner` for toast notifications during upload

### 3.2 Image Galleries/Grids with Previews

**Current state**: Custom `ImgStyleImageCard` + `ImageLightbox` pattern. Well-structured with consistent anatomy across upload, history, and admin pages.

**Available enhancements**:

1. **Aceternity `focus-cards`**: Hover to focus one card, blur others -- great gallery UX
2. **Aceternity `layout-grid`**: Animated grid with layout animations on click
3. **Aceternity `lens`**: Image zoom on hover for thumbnails
4. **Magic UI `lens`**: Alternative image zoom implementation
5. **Aceternity `apple-cards-carousel`**: Sleek carousel for featured images
6. **Aceternity `images-slider`**: Full-page image slider with keyboard nav
7. **shadcn `aspect-ratio`**: Maintain consistent thumbnail aspect ratios
8. **shadcn `carousel`** (embla-carousel-react): Accessible image carousel

### 3.3 Copy-to-Clipboard Functionality

**Current state**: Custom `CopyButton.tsx` using `navigator.clipboard.writeText()` + `react-hot-toast`. Works well with URL, Markdown, and BBcode formats.

**Available options**:

1. **Magic UI `script-copy-btn`**: Pre-built copy button with animation
2. **shadcn `tooltip`**: Add "Copied!" tooltip feedback instead of or in addition to toast
3. **shadcn `sonner`**: Replace `react-hot-toast` with shadcn's recommended toast library (sonner). The official sonner component auto-integrates with `next-themes` for dark mode and uses CSS variables for theming.

### 3.4 Progress Indicators

**Current state**: Custom `UploadProgressCard` with inline progress bar (div-based, ARIA progressbar role).

**Available options**:

1. **shadcn `progress`**: Official progress bar using `@radix-ui/react-progress`. Clean, accessible, CSS-variable themed.
2. **shadcn `spinner`**: Official spinner component using `class-variance-authority`.
3. **Aceternity `multi-step-loader`**: Multi-step loading indicator for complex upload flows (compress -> upload -> process).
4. **Aceternity `loader`**: Collection of minimal loaders.
5. **Magic UI `animated-circular-progress-bar`**: Circular progress with percentage -- good for per-image upload progress.
6. **shadcn `skeleton`**: Skeleton loading placeholders for image cards while thumbnails load.

### 3.5 Empty States

**Current state**: Custom `ImgGalleryEmptyState` component with icon, title, and optional description.

**Available options**:

1. **shadcn `empty`**: Official empty state component with `Empty`, `EmptyHeader`, `EmptyIcon`, `EmptyTitle`, `EmptyDescription`, `EmptyActions` sub-components. More structured and composable than the current custom implementation.

### 3.6 Skeleton Loaders

**Current state**: A `.skeleton-glass` CSS utility class in `globals.css` (`animate-pulse rounded-md bg-muted`).

**Available options**:

1. **shadcn `skeleton`**: Official component: `animate-pulse rounded-md bg-accent`. Simple, one-line. Can wrap it for image-card-shaped skeletons.

### 3.7 Toast Notifications

**Current state**: `react-hot-toast` (v2.5.2) -- already in `package.json`.

**shadcn recommendation**: The official shadcn/ui v4 ships with **sonner** as its toast library. The `sonner` component:
- Integrates with `next-themes` for automatic dark mode
- Uses CSS variables (`--normal-bg`, `--normal-text`, `--normal-border`) mapped to shadcn tokens
- Provides rich toast types: success, error, warning, info, loading
- Has a smaller bundle than react-hot-toast
- Is the "official" choice for new shadcn/ui projects

**Migration consideration**: The project uses `react-hot-toast` extensively. Switching to sonner would require:
- Replacing `toast.success()`, `toast.error()` calls
- Updating the `Toaster` component in layout
- Sonner has a similar API so migration is straightforward

---

## 4. Theming Best Practices

### 4.1 Current Project Theme System

The project uses a custom CSS variable system in `frontend/src/app/globals.css`:

```css
:root {
  --background: var(--color-surface);       /* 248 250 252 (slate-50) */
  --foreground: var(--color-ink);           /* 15 23 42 (slate-900) */
  --card: var(--color-panel);               /* 255 255 255 */
  --primary: var(--color-accent);           /* 124 58 237 (violet-600) */
  --primary-foreground: var(--color-accent-foreground); /* 255 255 255 */
  --secondary: var(--color-accent-soft);    /* 237 233 254 (violet-100) */
  --muted: 241 245 249;                     /* slate-100 */
  --border: var(--color-border);            /* 203 213 225 (slate-300) */
  --ring: var(--color-accent);              /* 124 58 237 */
  --radius: 0.5rem;
  /* ... plus destructive, accent, popover, input, etc. */
}

:root[data-theme="dark"], .dark {
  /* Dark mode overrides with violet accent */
}
```

**Key observations**:
- Uses raw RGB values (not HSL) -- shadcn/ui v4 standardizes on HSL
- Has an indirection layer (`--color-surface`, `--color-ink`, etc.) that maps to shadcn tokens
- Dark mode triggered by `data-theme="dark"` attribute or `.dark` class
- Custom scrollbar styling
- Custom focus-visible ring
- Utility classes: `.glass-panel`, `.gradient-text`, `.eyebrow-label`, `.toolbar-surface`, `.gallery-grid`, `.table-surface`, `.link-underline`, `.skeleton-glass`

### 4.2 shadcn/ui v4 Theming Standard

shadcn/ui v4 uses **HSL color space** with CSS variables. The standard token set:

```css
:root {
  --background: 0 0% 100%;          /* HSL: white */
  --foreground: 240 10% 3.9%;       /* HSL: near-black */
  --card: 0 0% 100%;
  --card-foreground: 240 10% 3.9%;
  --popover: 0 0% 100%;
  --popover-foreground: 240 10% 3.9%;
  --primary: 240 5.9% 10%;
  --primary-foreground: 0 0% 98%;
  --secondary: 240 4.8% 95.9%;
  --secondary-foreground: 240 5.9% 10%;
  --muted: 240 4.8% 95.9%;
  --muted-foreground: 240 3.8% 46.1%;
  --accent: 240 4.8% 95.9%;
  --accent-foreground: 240 5.9% 10%;
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 0 0% 98%;
  --border: 240 5.9% 90%;
  --input: 240 5.9% 90%;
  --ring: 240 5.9% 10%;
  --radius: 0.5rem;
}

.dark {
  --background: 240 10% 3.9%;
  --foreground: 0 0% 98%;
  /* ... dark variants ... */
}
```

**Key differences from current project**:
1. **HSL vs RGB**: shadcn v4 uses HSL (e.g., `240 5.9% 10%`), project uses RGB (e.g., `15 23 42`)
2. **No indirection**: shadcn uses tokens directly; project has `--color-*` indirection
3. **Dark mode**: shadcn uses `.dark` class; project uses `[data-theme="dark"]` or `.dark`
4. **Color values**: shadcn themes are neutral (zinc/slate/gray/stone); project uses violet accent

### 4.3 Migration Path (if desired)

To align with shadcn/ui v4 theming:

1. **Switch to HSL**: Convert all RGB values to HSL. The current violet accent `124 58 237` in HSL is approximately `271 81% 58%`.
2. **Remove indirection**: Map tokens directly (remove `--color-surface` etc.) or keep indirection but use HSL values.
3. **Use shadcn theme as base**: Start from `theme-zinc` or `theme-slate`, then customize `--primary` to violet.
4. **Dark mode**: The project already supports `.dark` class, which is compatible.
5. **CSS file**: shadcn v4 uses `@import "tailwindcss"` instead of `@tailwind base/components/utilities` (Tailwind CSS v4 syntax). The project uses Tailwind v3 (`^3.4.17`), so this would require a Tailwind upgrade.

### 4.4 Color System Recommendations

For an image hosting app, the current violet accent scheme works well. Key considerations:

- **Neutral backgrounds**: Images look best on neutral backgrounds (white/light gray or dark gray). The current `--color-surface: 248 250 252` (slate-50) is good.
- **Accent color**: Violet/purple is distinctive and works for a tech product. Keep it.
- **Dark mode**: Essential for an image-focused app -- many users prefer dark mode for viewing images. The current dark mode implementation is solid.
- **Border radius**: `--radius: 0.5rem` (8px) is standard. shadcn v4 themes vary: zinc/slate/gray use `0.5rem`, stone uses `0.95rem`.

### 4.5 Dark Mode Implementation

Current approach (good, keep):
- `data-theme="dark"` attribute on `<html>` for theme switching
- `.dark` class as fallback selector
- `color-scheme: dark` for native browser elements
- `UiPreferenceSync` component for persistence

shadcn v4 standard: Uses `next-themes` package with `<ThemeProvider>`. The project doesn't use `next-themes` currently. If adopting shadcn `sonner` for toasts, `next-themes` becomes a dependency.

---

## 5. Dependency Analysis

### 5.1 Current Dependencies (from package.json)

```
@radix-ui/react-dialog, @radix-ui/react-dropdown-menu, @radix-ui/react-label,
@radix-ui/react-select, @radix-ui/react-separator, @radix-ui/react-slot,
@radix-ui/react-tabs, class-variance-authority, clsx, lucide-react,
next, react, react-dom, react-hot-toast, tailwind-merge, zustand
```

### 5.2 New Dependencies Needed for Recommended Components

**If adding shadcn `progress`**: `@radix-ui/react-progress`
**If adding shadcn `tooltip`**: `@radix-ui/react-tooltip`
**If adding shadcn `popover`**: `@radix-ui/react-popover`
**If adding shadcn `scroll-area`**: `@radix-ui/react-scroll-area`
**If adding shadcn `avatar`**: `@radix-ui/react-avatar`
**If adding shadcn `checkbox`**: `@radix-ui/react-checkbox`
**If adding shadcn `sonner`**: `sonner`, `next-themes`
**If adding shadcn `skeleton`**: none (zero-dependency)
**If adding shadcn `empty`**: none (zero-dependency)
**If adding shadcn `pagination`**: none (zero-dependency)
**If adding shadcn `spinner`**: `class-variance-authority` (already have)
**If adding shadcn `alert-dialog`**: `@radix-ui/react-alert-dialog`
**If adding shadcn `aspect-ratio`**: `@radix-ui/react-aspect-ratio`
**If adding shadcn `toggle`**: `@radix-ui/react-toggle`
**If adding shadcn `switch`**: `@radix-ui/react-switch`
**If adding shadcn `accordion`**: `@radix-ui/react-accordion`
**If adding shadcn `collapsible`**: `@radix-ui/react-collapsible`
**If adding shadcn `hover-card`**: `@radix-ui/react-hover-card`
**If adding shadcn `carousel`**: `embla-carousel-react`
**If adding shadcn `drawer`**: `vaul`
**If adding shadcn `command`**: `cmdk`
**If adding shadcn `form`**: `react-hook-form`, `zod`, `@hookform/resolvers`
**If adding shadcn `sidebar`**: (uses existing deps)
**If adding shadcn `resizable`**: `react-resizable-panels`
**If adding Aceternity components**: `motion` (Framer Motion) -- most aceternity components depend on it
**If adding Aceternity `file-upload`**: `react-dropzone`, `motion`, `@tabler/icons-react`
**If adding Magic UI components**: `motion` (framer-motion) -- many magic UI components depend on it

---

## 6. Summary: Recommended Component Adoption Priority

### Tier 1 -- Immediate Value, Zero/Low Dependency Cost

1. **skeleton** -- zero deps, replaces `.skeleton-glass` utility
2. **empty** -- zero deps, structured empty states
3. **spinner** -- uses existing `class-variance-authority`
4. **pagination** -- zero deps, needed for admin image list
5. **kbd** -- zero deps, keyboard shortcut hints

### Tier 2 -- High UX Value, One New Dep Each

6. **progress** -- `@radix-ui/react-progress`, replaces custom progress bar
7. **tooltip** -- `@radix-ui/react-tooltip`, enhances copy buttons and image info
8. **popover** -- `@radix-ui/react-popover`, rich hover previews
9. **scroll-area** -- `@radix-ui/react-scroll-area`, consistent scrollbars
10. **avatar** -- `@radix-ui/react-avatar`, admin user display
11. **alert-dialog** -- `@radix-ui/react-alert-dialog`, delete confirmations
12. **aspect-ratio** -- `@radix-ui/react-aspect-ratio`, consistent thumbnails

### Tier 3 -- Major Feature Enablement

13. **sonner** -- `sonner` + `next-themes`, replaces react-hot-toast
14. **checkbox** -- `@radix-ui/react-checkbox`, batch image selection
15. **toggle / toggle-group** -- `@radix-ui/react-toggle`, view mode switching
16. **switch** -- `@radix-ui/react-switch`, settings toggles

### Tier 4 -- Third-Party Enhancement (Optional)

17. **Aceternity `file-upload`** -- if replacing custom dropzone
18. **Aceternity `lens`** or **Magic UI `lens`** -- image zoom on hover
19. **Aceternity `focus-cards`** -- gallery hover effect
20. **Magic UI `confetti`** -- upload success celebration
21. **Magic UI `script-copy-btn`** -- enhanced copy button

---

## Caveats / Not Found

- The shadcn/ui official website (ui.shadcn.com) is an SPA and does not serve raw JSON from its API endpoints directly -- the registry data was fetched from the GitHub repository (`shadcn-ui/ui`) raw files instead.
- Magic UI's website (magicui.design) is also an SPA; component data was fetched from the GitHub raw registry at `magicuidesign/magicui`.
- The project currently uses **Tailwind CSS v3** (`^3.4.17`). shadcn/ui v4 components are designed for Tailwind CSS v4 with the new `@import "tailwindcss"` syntax and `data-slot` attributes. Using shadcn v4 components with Tailwind v3 may require adaptation (removing `data-slot` attributes, adjusting CSS imports). The project should consider upgrading to Tailwind v4 for full compatibility.
- The project's custom CSS variable system uses RGB values while shadcn v4 uses HSL. This is not a blocking issue -- both work -- but mixing conventions could cause confusion.
- `react-hot-toast` vs `sonner`: Both are excellent toast libraries. The project already has `react-hot-toast` integrated. Switching to `sonner` is optional and primarily beneficial for tighter shadcn integration and `next-themes` dark mode support.
- The `@better-upload` registry was listed in the directory but its specific components were not fetched in detail.