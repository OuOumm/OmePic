# Product Intro Page PRD

## Goal

Create a standalone product introduction page at `test.html` in the repo root.

## Scope

- Replace the current empty `test.html` with a complete static HTML page.
- Keep everything self-contained in the single file.
- Do not introduce a framework, build step, or external asset dependency.

## Requirements

- The page must feel product-grade rather than like a template.
- The visual direction should be bright, eye-catching, and polished.
- Include noticeable but controlled motion.
- Include glow, light, or atmospheric effects.
- Create a sense of depth with layered surfaces, shadows, gradients, and/or perspective.
- The layout must work on desktop and mobile.
- Content language should be Chinese.

## Content Structure

- Hero section with strong headline, short supporting copy, and clear CTA buttons.
- Product value section with 3 to 4 key highlights.
- Visual showcase or feature cards that sell speed, stability, and image quality.
- Simple trust or metrics section.
- Closing CTA section.

## Visual Thesis

- Route: Generation
- Change intensity: full rebuild
- Thesis: bright warm neon product landing page with glossy depth, glass-like surfaces, orange/coral/gold accents, soft radial lighting, and restrained motion that improves hierarchy.

## Technical Constraints

- Use semantic HTML.
- Use internal `<style>` and optional lightweight inline JavaScript only if needed for presentation.
- Prefer maintainable CSS variables for theme colors and effects.
- Avoid excessive animation loops that make the page noisy.
- No placeholder lorem ipsum.

## Acceptance

- `test.html` opens directly in a browser and renders a complete landing page.
- The page includes animation and lighting effects without harming readability.
- The page has obvious visual hierarchy and dimensionality.
- The result is meaningfully more distinctive than a generic SaaS landing page.
