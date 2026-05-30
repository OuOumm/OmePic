# shadcn/ui Official Style Notes

## Sources

* shadcn/ui Card docs: https://ui.shadcn.com/docs/components/card
* shadcn/ui Blocks overview: https://ui.shadcn.com/blocks
* shadcn/ui Sidebar blocks: https://ui.shadcn.com/blocks/sidebar

## Findings

* Official card usage is composition-based: `Card`, `CardHeader`, `CardTitle`, `CardDescription`, optional action, `CardContent`, and `CardFooter`. This supports clear hierarchy without custom wrapper styles per page.
* The current official card docs include compact sizing and image examples, which maps well to dense app panels, recent uploads, image previews, and admin sections.
* Official blocks emphasize product-app layouts: sidebar/provider/inset shells, sticky or stable headers, separators, section cards, data tables, muted placeholder panels, and tight page padding.
* The official block language is clean and neutral rather than decorative: borders, muted backgrounds, small radius, compact spacing, and visible component states do most of the work.

## Repo Mapping

* Preserve `frontend/components.json` conventions: `new-york`, `slate`, CSS variables, lucide icons.
* Use shared primitives to carry the visual system instead of rebuilding per-page glass panels.
* Public pages should feel like an application workspace, not a marketing landing page.
* Admin pages should align with dashboard/sidebar block patterns while using the existing admin session and route structure.
