type DialogOptions = {
  onClose?: () => void;
};

const focusableSelector = [
  'a[href]',
  'button:not([disabled])',
  'textarea:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

function getFocusableElements(node: HTMLElement) {
  return Array.from(node.querySelectorAll<HTMLElement>(focusableSelector)).filter((element) => !element.hasAttribute('disabled') && !element.getAttribute('aria-hidden'));
}

export function accessibleDialog(node: HTMLElement, options: DialogOptions = {}) {
  const previouslyFocused = document.activeElement instanceof HTMLElement ? document.activeElement : null;

  function focusInitialElement() {
    const firstFocusable = getFocusableElements(node)[0];
    (firstFocusable ?? node).focus();
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      options.onClose?.();
      return;
    }

    if (event.key !== 'Tab') return;

    const focusable = getFocusableElements(node);
    if (!focusable.length) {
      event.preventDefault();
      node.focus();
      return;
    }

    const first = focusable[0];
    const last = focusable[focusable.length - 1];

    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  node.addEventListener('keydown', handleKeydown);
  requestAnimationFrame(focusInitialElement);

  return {
    update(nextOptions: DialogOptions = {}) {
      options = nextOptions;
    },
    destroy() {
      node.removeEventListener('keydown', handleKeydown);
      previouslyFocused?.focus();
    },
  };
}
