<script lang="ts">
  import { page } from '$app/state';
  import { Menu, Moon, Sun, Languages, Upload, History, KeyRound, Shield } from 'lucide-svelte';
  import ToastViewport from './ToastViewport.svelte';
  import { preferences, setLanguage, setTheme, resolvedTheme } from '@/stores/preferences.svelte';
  import { t } from '@/i18n';
  import { initialThemeScript } from '@/utils';

  let menuOpen = $state(false);

  const siteName = $derived(preferences.runtimeSettings?.site.name || 'OmePic');

  const nav = $derived([
    { href: '/', label: t(preferences.language, 'nav.upload'), icon: Upload },
    { href: '/history', label: t(preferences.language, 'nav.history'), icon: History },
    { href: '/api', label: t(preferences.language, 'nav.api'), icon: KeyRound },
    { href: '/admin/dashboard', label: t(preferences.language, 'nav.admin'), icon: Shield },
  ]);
  const mobileNavId = 'site-mobile-navigation';

  let { children } = $props();

  const currentTheme = $derived(preferences.theme);
  const scriptTag = 'script';
  const themeBootstrapScript = initialThemeScript();

  function isActive(href: string) {
    return href === '/' ? page.url.pathname === '/' : page.url.pathname.startsWith(href);
  }

  function applyTheme() {
    if (typeof document === 'undefined') return;
    const isDark = resolvedTheme() === 'dark';
    if (document.documentElement.classList.contains('dark') === isDark) return;
    document.documentElement.classList.toggle('dark', isDark);
  }

  function toggleTheme() {
    setTheme(resolvedTheme() === 'dark' ? 'light' : 'dark');
  }

  $effect(() => {
    if (currentTheme) applyTheme();
    if (currentTheme !== 'system' || typeof window === 'undefined') return;

    const media = window.matchMedia('(prefers-color-scheme: dark)');
    media.addEventListener('change', applyTheme);
    return () => media.removeEventListener('change', applyTheme);
  });
</script>

<svelte:head>
  <svelte:element this={scriptTag}>{themeBootstrapScript}</svelte:element>
</svelte:head>

<svelte:window onstorage={applyTheme} />

<div class="min-h-screen">
  <header class="sticky top-0 z-50 border-b-2 ink-line bg-[hsl(var(--paper)/0.86)] backdrop-blur-md will-change-[backdrop-filter]">
    <div class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 lg:px-6">
      <a href="/" class="group flex items-center gap-3 font-black tracking-tight">
        <span class="grid size-10 place-items-center border-2 ink-line bg-[hsl(var(--marker-yellow))] shadow-[4px_4px_0_hsl(var(--ink))] transition-transform group-hover:-rotate-3">OP</span>
        <span class="text-xl">{siteName}</span>
      </a>

      <nav class="hidden items-center gap-1 lg:flex">
        {#each nav as item (item.href)}
          <a class="flex items-center gap-2 px-3 py-2 text-sm font-bold hover:marker-highlight" href={item.href} aria-current={isActive(item.href) ? 'page' : undefined}>
            <item.icon class="size-4" aria-hidden="true" />
            {item.label}
          </a>
        {/each}
      </nav>

      <div class="hidden items-center gap-2 lg:flex">
        <button class="studio-button text-xs" type="button" onclick={() => setLanguage(preferences.language === 'zh' ? 'en' : 'zh')}>
          <Languages class="size-4" />
          {preferences.language === 'zh' ? 'EN' : '中文'}
        </button>
        <button class="studio-button text-xs" type="button" onclick={toggleTheme}>
          {#if resolvedTheme() === 'dark'}<Sun class="size-4" />{:else}<Moon class="size-4" />{/if}
          {resolvedTheme() === 'dark' ? t(preferences.language, 'common.light') : t(preferences.language, 'common.dark')}
        </button>
      </div>

      <button class="studio-button lg:hidden" type="button" onclick={() => (menuOpen = !menuOpen)} aria-label={t(preferences.language, 'nav.menu')} aria-expanded={menuOpen} aria-controls={mobileNavId}>
        <Menu class="size-4" aria-hidden="true" />
      </button>
    </div>

    {#if menuOpen}
      <nav id={mobileNavId} class="grid border-t-2 ink-line bg-[hsl(var(--paper))] px-4 py-3 lg:hidden">
        {#each nav as item (item.href)}
          <a class="flex items-center gap-2 py-3 font-bold" href={item.href} onclick={() => (menuOpen = false)} aria-current={isActive(item.href) ? 'page' : undefined}>
            <item.icon class="size-4" aria-hidden="true" />
            {item.label}
          </a>
        {/each}
      </nav>
    {/if}
  </header>

  <main class="mx-auto w-full max-w-7xl px-4 py-6 lg:px-6 lg:py-8">
    {@render children()}
  </main>
</div>

<ToastViewport />

