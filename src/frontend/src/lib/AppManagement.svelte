<script lang="ts">
import { onMount } from 'svelte';
import AppBlocklist from './AppBlocklist.svelte';
import AppLeaderboard from './AppLeaderboard.svelte';
import AppSearch from './AppSearch.svelte';

let activeTab: 'leaderboard' | 'search' | 'blocklist' = 'leaderboard';
function showSubView(view: 'leaderboard' | 'search' | 'blocklist') {
  activeTab = view;
}

onMount(() => {
  // Initial load for leaderboard
  // This will be handled by the component itself
});
</script>

<div id="app-management-view">
  <!-- Tab Navigation -->
  <ul class="nav nav-tabs" id="appManTabs" role="tablist">
    <li class="nav-item" role="presentation">
      <button
        class="nav-link"
        class:active={activeTab === 'leaderboard'}
        id="leaderboard-tab"
        type="button"
        role="tab"
        on:click={() => showSubView('leaderboard')}
      >
        Bảng xếp hạng
      </button>
    </li>
    <li class="nav-item" role="presentation">
      <button
        class="nav-link"
        class:active={activeTab === 'search'}
        id="search-tab"
        type="button"
        role="tab"
        on:click={() => showSubView('search')}
      >
        Tìm kiếm & Chặn ứng dụng
      </button>
    </li>
    <li class="nav-item" role="presentation">
      <button
        class="nav-link"
        class:active={activeTab === 'blocklist'}
        id="blocklist-tab"
        type="button"
        role="tab"
        on:click={() => showSubView('blocklist')}
      >
        Quản lý danh sách chặn
      </button>
    </li>
  </ul>

  <!-- Tab Content -->
  <div class="tab-content" id="appManTabsContent">
    {#if activeTab === 'leaderboard'}
      <div id="leaderboard-view" role="tabpanel">
        <AppLeaderboard />
      </div>
    {:else if activeTab === 'search'}
      <div id="search-view" role="tabpanel">
        <AppSearch />
      </div>
    {:else if activeTab === 'blocklist'}
      <div id="blocklist-view" role="tabpanel">
        <AppBlocklist />
      </div>
    {/if}
  </div>
</div>
