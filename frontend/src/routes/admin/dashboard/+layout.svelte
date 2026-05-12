<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { BarChart3, Gauge, Image, LogOut, Megaphone, Server, Settings, ShieldAlert, ShieldCheck, SlidersHorizontal } from 'lucide-svelte';
  import { adminGetStatus } from '@/api';
  import { t } from '@/i18n';
  import { clearAdminToken, preferences } from '@/stores/preferences.svelte';
  import { isAbortError } from '@/utils';
  import type { AdminStatus } from '@/types';

  let status = $state<AdminStatus | null>(null);
  let { children } = $props();

  const links = $derived([
    { href: '/admin/dashboard', label: t(preferences.language, 'admin.sidebarStatus'), icon: BarChart3 },
    { href: '/admin/dashboard/images', label: t(preferences.language, 'admin.sidebarImages'), icon: Image },
    { href: '/admin/dashboard/security', label: t(preferences.language, 'admin.sidebarSecurity'), icon: ShieldAlert },
    { href: '/admin/dashboard/settings', label: t(preferences.language, 'admin.sidebarSettings'), icon: Settings },
  ]);
  const settingsTabs = $derived([
    { href: '/admin/dashboard/settings?tab=runtime', tab: 'runtime', label: t(preferences.language, 'admin.submenuRuntime'), icon: SlidersHorizontal },
    { href: '/admin/dashboard/settings?tab=storage', tab: 'storage', label: t(preferences.language, 'admin.submenuStorage'), icon: Server },
    { href: '/admin/dashboard/settings?tab=announcements', tab: 'announcements', label: t(preferences.language, 'admin.submenuAnnouncements'), icon: Megaphone }
  ]);
  const securityTabs = $derived([
    { href: '/admin/dashboard/security?tab=abuse', tab: 'abuse', label: t(preferences.language, 'admin.submenuAbuse'), icon: ShieldCheck },
    { href: '/admin/dashboard/security?tab=rate-limit', tab: 'rate-limit', label: t(preferences.language, 'admin.submenuRateLimit'), icon: Gauge }
  ]);
  const isSettingsPage = $derived(page.url.pathname === '/admin/dashboard/settings');
  const isSecurityPage = $derived(page.url.pathname === '/admin/dashboard/security');
  const isDashboardEntry = $derived(page.url.pathname === '/admin/dashboard');
  const activeSettingsTab = $derived(page.url.searchParams.get('tab') ?? 'runtime');
  const activeSecurityTab = $derived(page.url.searchParams.get('tab') ?? 'abuse');

  function isActiveLink(href: string) {
    return href === '/admin/dashboard' ? page.url.pathname === href : page.url.pathname.startsWith(href);
  }

  function logout() {
    clearAdminToken();
    void goto('/admin/dashboard');
  }

  $effect(() => {
    if (!preferences.adminToken && !isDashboardEntry) {
      void goto('/admin/dashboard', { replaceState: true });
    }
  });

  $effect(() => {
    if (!preferences.adminToken) {
      status = null;
      return;
    }
    const controller = new AbortController();
    adminGetStatus(preferences.adminToken, controller.signal)
      .then((next) => {
        if (!controller.signal.aborted) status = next;
      })
      .catch((err) => {
        if (!isAbortError(err)) {
          clearAdminToken();
          void goto('/admin/dashboard', { replaceState: true });
        }
      });
    return () => controller.abort();
  });
</script>

<div class={preferences.adminToken ? 'grid gap-6 lg:grid-cols-[250px_1fr]' : 'grid min-h-[calc(100dvh-8rem)] place-items-center'}>
  {#if preferences.adminToken}
    <aside class="studio-panel h-fit p-4 lg:sticky lg:top-24">
      <div class="border-b-2 ink-line pb-3">
        <h1 class="text-2xl font-black">{t(preferences.language, 'admin.blueprintTitle')}</h1>
      </div>
      <nav class="mt-4" aria-label={t(preferences.language, 'admin.blueprintTitle')}>
        <ul class="grid gap-2">
          {#each links as item (item.href)}
            <li>
              <a class="flex min-w-0 items-center gap-2 border-b-2 border-dashed border-[hsl(var(--ink)/0.24)] py-2 font-black hover:marker-highlight focus-visible:marker-highlight" href={item.href} aria-current={isActiveLink(item.href) ? 'page' : undefined}>
                <item.icon class="size-4 shrink-0" aria-hidden="true" />
                <span class="min-w-0 truncate">{item.label}</span>
              </a>
              {#if item.href === '/admin/dashboard/security' && isSecurityPage}
                <ul class="ml-6 grid gap-1 border-b-2 border-dashed border-[hsl(var(--ink)/0.24)] pb-2" aria-label={t(preferences.language, 'admin.sidebarSecurity')}>
                  {#each securityTabs as tab (tab.tab)}
                    <li>
                      <a class="flex min-w-0 items-center gap-2 px-2 py-1.5 text-sm font-black {activeSecurityTab === tab.tab ? 'marker-highlight' : 'text-[hsl(var(--ink-muted))] hover:marker-highlight focus-visible:marker-highlight'}" href={tab.href} aria-current={activeSecurityTab === tab.tab ? 'page' : undefined}>
                        <tab.icon class="size-3.5 shrink-0" aria-hidden="true" />
                        <span class="min-w-0 truncate">{tab.label}</span>
                      </a>
                    </li>
                  {/each}
                </ul>
              {/if}
              {#if item.href === '/admin/dashboard/settings' && isSettingsPage}
                <ul class="ml-6 grid gap-1 border-b-2 border-dashed border-[hsl(var(--ink)/0.24)] pb-2" aria-label={t(preferences.language, 'admin.sidebarSettings')}>
                  {#each settingsTabs as tab (tab.tab)}
                    <li>
                      <a class="flex min-w-0 items-center gap-2 px-2 py-1.5 text-sm font-black {activeSettingsTab === tab.tab ? 'marker-highlight' : 'text-[hsl(var(--ink-muted))] hover:marker-highlight focus-visible:marker-highlight'}" href={tab.href} aria-current={activeSettingsTab === tab.tab ? 'page' : undefined}>
                        <tab.icon class="size-3.5 shrink-0" aria-hidden="true" />
                        <span class="min-w-0 truncate">{tab.label}</span>
                      </a>
                    </li>
                  {/each}
                </ul>
              {/if}
            </li>
          {/each}
        </ul>
      </nav>
      {#if status}
        <dl class="mt-5 grid gap-2 text-sm">
          <div class="flex justify-between"><dt>{t(preferences.language, 'admin.totalImages')}</dt><dd class="font-black">{status.total_images}</dd></div>
          <div class="flex justify-between"><dt>{t(preferences.language, 'admin.todayUploads')}</dt><dd class="font-black">{status.today_uploads}</dd></div>
          <div class="flex justify-between"><dt>{t(preferences.language, 'admin.uniqueTokens')}</dt><dd class="font-black">{status.unique_tokens}</dd></div>
        </dl>
      {/if}
      <button class="studio-button mt-5 w-full" type="button" onclick={logout}><LogOut class="size-4" />{t(preferences.language, 'admin.logout')}</button>
    </aside>
  {/if}

  {#if preferences.adminToken || isDashboardEntry}
    <section class={preferences.adminToken ? 'min-w-0 overflow-hidden' : 'w-full max-w-[520px]'}>
      {@render children()}
    </section>
  {/if}
</div>
