<script lang="ts">
import { onMount } from 'svelte';
import { openConfirmModal } from './modalStore';
import { showToast } from './toastStore';

let isAutostartEnabled = false;
let autostartToggleBtnDisabled = false;

async function loadAutostartStatus(): Promise<void> {
  try {
    isAutostartEnabled = await window.go.main.App.GetAutostartStatus();
    autostartToggleBtnDisabled = false;
  } catch (error) {
    console.error('Error loading autostart status:', error);
    showToast('Không hỗ trợ tự động khởi động trên HĐH này', 'info');
    autostartToggleBtnDisabled = true;
  }
}

async function toggleAutostart(): Promise<void> {
  autostartToggleBtnDisabled = true;
  try {
    if (isAutostartEnabled) {
      await window.go.main.App.DisableAutostart();
    } else {
      await window.go.main.App.EnableAutostart();
    }
    showToast(
      isAutostartEnabled
        ? 'Đã tắt tự động khởi động.'
        : 'Đã bật tự động khởi động.',
      'success',
    );
    loadAutostartStatus(); // Refresh status after action
  } catch (e) {
    console.error('Error toggling autostart:', e);
    showToast(
      `Đã xảy ra lỗi: ${e instanceof Error ? e.message : 'Unknown error'}`,
      'error',
    );
  } finally {
    autostartToggleBtnDisabled = false;
  }
}

async function clearAppHistory(): Promise<void> {
  openConfirmModal('Xóa lịch sử ứng dụng', 'clearAppHistory');
}

async function clearWebHistory(): Promise<void> {
  openConfirmModal('Xóa lịch sử Web', 'clearWebHistory');
}

onMount(() => {
  loadAutostartStatus();
});
</script>

<div id="settings-view">
  <div class="row">
    <div class="col-md-8 mx-auto">
      <!-- Autostart Settings Card -->
      <div class="card mb-4">
        <div class="card-header">
          <h4>Khởi động cùng Windows</h4>
        </div>
        <div class="card-body">
          <p class="card-text">
            Trạng thái:
            <span
              class="badge {isAutostartEnabled ? 'bg-success' : 'bg-secondary'}"
              >{isAutostartEnabled ? 'Đã bật' : 'Đã tắt'}</span
            >
          </p>
          <p class="card-text">
            Bật tùy chọn này để ProcGuard tự động chạy khi bạn đăng nhập vào
            Windows.
          </p>
          <button
            id="autostart-toggle-btn"
            class="btn btn-primary"
            on:click={toggleAutostart}
            disabled={autostartToggleBtnDisabled}
          >
            {isAutostartEnabled
              ? 'Tắt tự động khởi động'
              : 'Bật tự động khởi động'}
          </button>
        </div>
      </div>

      <!-- History Management Card -->
      <div class="card mb-4">
        <div class="card-header">
          <h4>Quản lý lịch sử</h4>
        </div>
        <div class="card-body">
          <p class="card-text">
            Xóa dữ liệu thu thập được từ các ứng dụng và trang web.
          </p>
          <div class="alert alert-info py-2 small mb-3">
            <i class="bi bi-info-circle me-1"></i>
            Lưu ý: Một số ứng dụng chạy ngầm (như trình duyệt) có thể cần khởi động
            lại để được ghi nhận lại ngay sau khi xóa.
          </div>
          <div class="d-flex gap-2">
            <button class="btn btn-warning" on:click={clearAppHistory}>
              Xóa lịch sử ứng dụng
            </button>
            <button class="btn btn-warning" on:click={clearWebHistory}>
              Xóa lịch sử Web
            </button>
          </div>
        </div>
      </div>

      <!-- Uninstall Card -->
      <div class="card mb-4">
        <div class="card-header">
          <h4>Gỡ cài đặt ProcGuard</h4>
        </div>
        <div class="card-body">
          <p class="card-text">
            <b>Cảnh báo:</b> Thao tác này sẽ xóa toàn bộ dữ liệu và gỡ cài đặt ProcGuard
            khỏi hệ thống.
          </p>
          <button
            type="button"
            class="btn btn-danger"
            on:click={() =>
              openConfirmModal('Xác nhận gỡ cài đặt', 'uninstall')}
          >
            Gỡ cài đặt ProcGuard
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
