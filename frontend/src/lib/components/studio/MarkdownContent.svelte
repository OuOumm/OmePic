<script lang="ts">
  import type { Token, TokensList } from 'marked';

  type Props = {
    content: string;
    clamp?: boolean;
  };

  let { content, clamp = false }: Props = $props();
  let tokens = $state<TokensList | null>(null);
  let parsedContent = '';

  async function parseMarkdown(nextContent: string) {
    const [{ marked }, { default: DOMPurify }] = await Promise.all([
      import('marked'),
      import('dompurify'),
    ]);
    marked.use({ async: false, gfm: true, breaks: true });
    const nextTokens = marked.lexer(DOMPurify.sanitize(nextContent));
    if (parsedContent === nextContent) tokens = nextTokens;
  }

  $effect(() => {
    parsedContent = content;
    tokens = null;
    void parseMarkdown(content);
  });

  function textFromTokens(items: Token[] | TokensList | undefined) {
    if (!items) return '';
    return items.map((item) => ('text' in item ? item.text : '')).join('');
  }
</script>

<div class={['markdown-content text-[hsl(var(--ink-muted))]', clamp && 'line-clamp-2']}>
  {#each tokens ?? [] as token (token.raw)}
    {#if token.type === 'heading'}
      <svelte:element this={`h${Math.min(token.depth, 4)}`}>{textFromTokens(token.tokens)}</svelte:element>
    {:else if token.type === 'paragraph'}
      <p>{textFromTokens(token.tokens)}</p>
    {:else if token.type === 'list'}
      {#if token.ordered}
        <ol>
          {#each token.items as item (item.raw)}
            <li>{textFromTokens(item.tokens)}</li>
          {/each}
        </ol>
      {:else}
        <ul>
          {#each token.items as item (item.raw)}
            <li>{textFromTokens(item.tokens)}</li>
          {/each}
        </ul>
      {/if}
    {:else if token.type === 'blockquote'}
      <blockquote>{textFromTokens(token.tokens)}</blockquote>
    {:else if token.type === 'code'}
      <pre><code>{token.text}</code></pre>
    {:else if token.type === 'hr'}
      <hr />
    {:else if token.type !== 'space'}
      <p>{'text' in token ? token.text : token.raw}</p>
    {/if}
  {/each}
</div>

<style>
  .markdown-content :global(*) {
    overflow-wrap: anywhere;
  }

  .markdown-content :global(:first-child) {
    margin-top: 0;
  }

  .markdown-content :global(:last-child) {
    margin-bottom: 0;
  }

  .markdown-content :global(p) {
    margin: 0.65rem 0;
    font-weight: 600;
    line-height: 1.75;
  }

  .markdown-content :global(h1),
  .markdown-content :global(h2),
  .markdown-content :global(h3),
  .markdown-content :global(h4) {
    margin: 1rem 0 0.5rem;
    color: hsl(var(--ink));
    font-weight: 900;
    line-height: 1.2;
  }

  .markdown-content :global(h1) {
    font-size: 1.5rem;
  }

  .markdown-content :global(h2) {
    font-size: 1.3rem;
  }

  .markdown-content :global(h3) {
    font-size: 1.1rem;
  }

  .markdown-content :global(ul),
  .markdown-content :global(ol) {
    margin: 0.7rem 0;
    padding-left: 1.4rem;
    font-weight: 600;
    line-height: 1.7;
  }

  .markdown-content :global(ul) {
    list-style: disc;
  }

  .markdown-content :global(ol) {
    list-style: decimal;
  }

  .markdown-content :global(li + li) {
    margin-top: 0.25rem;
  }

  .markdown-content :global(blockquote) {
    margin: 0.85rem 0;
    border-left: 4px solid hsl(var(--ink));
    background: hsl(var(--paper-alt));
    padding: 0.65rem 0.85rem;
    font-weight: 700;
  }

  .markdown-content :global(code) {
    border: 2px solid hsl(var(--ink));
    background: hsl(var(--marker-yellow));
    padding: 0.05rem 0.25rem;
    color: hsl(var(--ink));
    font-weight: 900;
  }

  .markdown-content :global(pre) {
    margin: 0.85rem 0;
    overflow-x: auto;
    border: 2px solid hsl(var(--ink));
    background: hsl(var(--paper-alt));
    padding: 0.85rem;
  }

  .markdown-content :global(pre code) {
    border: 0;
    background: transparent;
    padding: 0;
  }

  .markdown-content :global(hr) {
    margin: 1rem 0;
    border: 0;
    border-top: 2px dashed hsl(var(--ink) / 0.32);
  }
</style>
