<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import SearchResultItem from './SearchResultItem.svelte';
  import { showToast } from './toastStore';
  import DateRangePicker from './DateRangePicker.svelte';

  interface SearchResultItem {
    processName: string;
    exePath: string;
    commercialName: string;
    icon: string;
    otherInfo: string;
  }

  let q = '';
  let searchResults = writable<SearchResultItem[]>([]);
  let selectedApps: string[] = [];
  let since: Date | null = null;
  let until: Date | null = new Date();

  function formatDateTime(date: Date | null): string {
    if (!date) return '';
    return date.toISOString();
  }

  async function performSearch(sinceStr: string, untilStr: string): Promise<void> {
    let url = '/api/search?q=' + encodeURIComponent(q);
    if (sinceStr) {
      url += '&since=' + encodeURIComponent(sinceStr);
    }
    if (untilStr) {
      url += '&until=' + encodeURIComponent(untilStr);
    }
    const res = await fetch(url, { cache: 'no-cache' });
    const data = await res.json();
    if (data && data.length > 0) {
      const items: SearchResultItem[] = await Promise.all(
        data.map(async (l: string[]) => {
          const processName = l[1];
          const exePath = l[4]; // exe_path is the 5th element
          let commercialName = '';
          let icon = '';

          if (exePath) {
            const appDetailsRes = await fetch(
              `/api/app-details?path=${encodeURIComponent(exePath)}`
            );
            if (appDetailsRes.ok) {
              const appDetails = await appDetailsRes.json();
              commercialName = appDetails.commercialName;
              icon = appDetails.icon;
            }
          }

          const otherInfo = l.filter((_, i) => i !== 1 && i !== 4).join(' | ');

          return {
            processName,
            exePath,
            commercialName,
            icon,
            otherInfo,
          };
        })
      );
      searchResults.set(items);
    } else {
      searchResults.set([]);
    }
  }

  function handleDateChange(event: CustomEvent<{ since: Date | null; until: Date | null }>) {
    since = event.detail.since;
    until = event.detail.until;
    performSearch(formatDateTime(since), formatDateTime(until));
  }

  async function block(): Promise<void> {
    if (selectedApps.length === 0) {
      showToast(
        'Vui lòng chọn một ứng dụng từ kết quả tìm kiếm để chặn.',
        'info'
      );
      return;
    }

    const uniqueApps = [...new Set(selectedApps)];

    const response = await fetch('/api/block', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ names: uniqueApps }),
    });

    if (!response.ok) {
      showToast(`Lỗi chặn ứng dụng: ${response.statusText}`, 'error');
      return;
    }

    showToast(`Đã chặn: ${uniqueApps.join(', ')}`, 'success');
    selectedApps = [];
  }

  onMount(() => {
    const now = new Date();
    since = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0);
    until = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59);
    performSearch(formatDateTime(since), formatDateTime(until));
  });
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Tìm kiếm Log ứng dụng</h5>
    <div class="input-group mb-3">
      <input
        type="text"
        class="form-control"
        id="q"
        placeholder="Nhập tên ứng dụng..."
        bind:value={q}
        on:input={() => performSearch(formatDateTime(since), formatDateTime(until))}
      />
      <button class="btn btn-danger" type="button" on:click={block}>
        Chặn mục đã chọn
      </button>
    </div>

    <div class="card mt-3">
      <div class="card-header d-flex align-items-center">
        <span class="me-3">Lọc theo thời gian định sẵn:</span>
        <div class="btn-group" role="group">
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() => performSearch('1 hour ago', 'now')}
          >
            1 giờ qua
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() => performSearch('24 hours ago', 'now')}
          >
            24 giờ qua
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            on:click={() => performSearch('7 days ago', 'now')}
          >
            7 ngày qua
          </button>
        </div>
      </div>
      <div class="card-body">
        <span class="me-3">Lọc theo thời gian bất kỳ:</span>
        <DateRangePicker {since} {until} on:change={handleDateChange} />
      </div>
    </div>

    <h5 class="mt-4">Kết quả tìm kiếm</h5>
    <div id="results" class="list-group">
      {#each $searchResults as item, i (item.processName + i)}
        <label class="list-group-item d-flex align-items-center">
          <input
            class="form-check-input me-2"
            type="checkbox"
            name="search-result-app"
            value={item.processName}
            bind:group={selectedApps}
          />
          <SearchResultItem {item} />
        </label>
      {:else}
        <div class="list-group-item">Không tìm thấy kết quả.</div>
      {/each}
    </div>
  </div>
</div>
