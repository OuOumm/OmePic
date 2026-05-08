export function clickOutside(node: HTMLElement, callback: () => void) {
  const handleClick = (event: MouseEvent) => {
    if (!node.contains(event.target as Node)) callback();
  };
  document.addEventListener('mousedown', handleClick);
  return {
    destroy() {
      document.removeEventListener('mousedown', handleClick);
    },
  };
}
