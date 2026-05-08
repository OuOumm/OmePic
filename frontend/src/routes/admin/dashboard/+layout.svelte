<script lang="ts">
  import { BarChart3, Image, LogOut, Settings, ShieldAlert } from 'lucide-svelte';
  import { adminGetStatus } from '@/api';
  import { t } from '@/i18n';
  import { clearAdminToken, preferences } from '@/stores/preferences.svelte';
  import type { AdminStatus } from '@/types';

  let status = $state<AdminStatus | null>(null);
  let { children } = $props();

  const links = $derived([
    { href: '/admin/dashboard', label: t(preferences.language, 'admin.sidebarStatus'), icon: BarChart3 },
    { href: '/admin/dashboard/images', label: t(preferences.language, 'admin.sidebarImages'), icon: Image },
    { href: '/admin/dashboard/security', label: t(preferences.language, 'admin.sidebarSecurity'), icon: ShieldAlert },
    { href: '/admin/dashboard/settings', label: t(preferences.language, 'admin.sidebarSettings'), icon: Settings },
  ]);

  function logout() {
    clearAdminToken();
    location.href = '/admin/dashboard';
  }

  $effect(() => {
    if (preferences.adminToken) {
      adminGetStatus(preferences.adminToken).then((next) => (status = next)).catch(() => clearAdminToken());
    }
  });
</script>

<div class="grid gap-6 lg:grid-cols-[250px_1fr]">
  <aside class="studio-panel h-fit p-4 lg:sticky lg:top-24">
    <div class="border-b-2 ink-line pb-3">
      <p class="text-xs font-black uppercase text-[hsl(var(--ink-muted))]">OmePic</p>
      <h1 class="text-2xl font-black">Admin Blueprint</h1>
    </div>
    <nav class="mt-4 grid gap-2">
      {#each links as item (item.href)}
        <a class="flex items-center gap-2 border-b-2 border-dashed border-[hsl(var(--ink)/0.24)] py-2 font-black hover:marker-highlight" href={item.href}>
          <item.icon class="size-4" />
          {item.label}
        </a>
      {/each}
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

  <section>
    {@render children()}
  </section>
</div>

