import { fromAction, type Attachment } from 'svelte/attachments';

/**
 * Moves modal overlay roots to <body> so viewport-fixed backdrops are not constrained
 * by route/layout containers with overflow, stacking, or containment styles.
 */
export function viewportPortal(node: HTMLElement) {
  if (typeof document === 'undefined' || !document.body) return {};

  const placeholder = document.createComment('omepic-viewport-portal');
  node.parentNode?.insertBefore(placeholder, node);
  document.body.appendChild(node);

  return {
    destroy() {
      node.remove();
      placeholder.remove();
    },
  };
}

export function attachViewportPortal(): Attachment<HTMLElement> {
  return fromAction(viewportPortal);
}
