<script lang="ts">
import { onMount } from 'svelte';
import { writable } from 'svelte/store';
import DateRangePicker from './DateRangePicker.svelte';
import SearchResultItem from './SearchResultItem.svelte';
import { showToast } from './toastStore';

interface SearchResultData {
  processName: string;
  exePath: string;
  commercialName: string;
  icon: string;
  otherInfo: string;
}

let q = '';
let searchResults = writable<SearchResultData[]>([]);
let selectedApps: string[] = [];
let since: Date | null = null;
let until: Date | null = new Date();

function formatDateTime(date: Date | null): string {
  if (!date) return '';
  return date.toISOString();
}

async function performSearch(
  sinceStr: string,
  untilStr: string,
): Promise<void> {
  try {
    console.log('Performing app search...', { sinceStr, untilStr, q });
    const data = await window.go.main.App.Search(q, sinceStr, untilStr);
    console.log('App search data received:', data);
    if (data && data.length > 0) {
      const items: SearchResultData[] = await Promise.all(
        data.map(async (l: string[]) => {
          const processName = l[1];
          const exePath = l[4]; // exe_path is the 5th element
          let commercialName = '';
          let icon = '';

          if (exePath) {
            try {
              const appDetails =
                await window.go.main.App.GetAppDetails(exePath);
              commercialName = appDetails.commercialName;
              icon = appDetails.icon;
            } catch (error) {
              console.error('Error fetching app details:', error);
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
        }),
      );
      searchResults.set(items);
    } else {
      searchResults.set([]);
    }
  } catch (error) {
    console.error('Error performing search:', error);
    searchResults.set([]);
  }
}

function handleDateChange(
  event: CustomEvent<{ since: Date | null; until: Date | null }>,
) {
  since = event.detail.since;
  until = event.detail.until;
  performSearch(formatDateTime(since), formatDateTime(until));
}

async function block(): Promise<void> {
  if (selectedApps.length === 0) {
    showToast(
      'Vui lòng chọn một ứng dụng từ kết quả tìm kiếm để chặn.',
      'info',
    );
    return;
  }

  const uniqueApps = [...new Set(selectedApps)];

  try {
    await window.go.main.App.BlockApps(uniqueApps);
    showToast(
      `Các ứng dụng đã chọn đã được thêm vào danh sách chặn`,
      'success',
    );
    selectedApps = [];
  } catch (error) {
    console.error('Error blocking apps:', error);
    showToast('Lỗi khi chặn ứng dụng.', 'error');
  }
}

onMount(() => {
  const now = new Date();
  since = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000); // 7 days ago
  until = new Date();
  performSearch(formatDateTime(since), formatDateTime(until));
});
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Lịch sử Ứng dụng</h5>
    <div class="input-group mb-3">
      <input
        type="text"
        class="form-control"
        id="q"
        placeholder="Nhập tên ứng dụng..."
        bind:value={q}
        on:input={() =>
          performSearch(formatDateTime(since), formatDateTime(until))}
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
