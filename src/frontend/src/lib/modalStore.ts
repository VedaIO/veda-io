import { get, writable } from 'svelte/store';
import { showToast } from './toastStore';

export type ConfirmAction = 'uninstall' | 'clearAppHistory' | 'clearWebHistory';

export const isConfirmModalOpen = writable(false);
export const confirmModalPassword = writable('');
export const confirmModalError = writable('');
export const confirmModalTitle = writable('');
export const confirmModalAction = writable<ConfirmAction | null>(null);

/**
 * Opens the confirmation modal with a specific title and action.
 * @param title The title to display in the modal.
 * @param action The internal action to perform on success.
 */
export function openConfirmModal(title: string, action: ConfirmAction) {
  confirmModalPassword.set('');
  confirmModalError.set('');
  confirmModalTitle.set(title);
  confirmModalAction.set(action);
  isConfirmModalOpen.set(true);
}

/**
 * Handles the submission of the confirmation modal.
 * Verifies the password and executes the selected action.
 */
export async function handleConfirmSubmit() {
  const password = get(confirmModalPassword);
  const action = get(confirmModalAction);

  if (!action) return;

  try {
    switch (action) {
      case 'uninstall':
        await window.go.main.App.Uninstall(password);
        isConfirmModalOpen.set(false);
        // Give the modal a moment to close before closing the page
        setTimeout(() => {
          window.location.href = 'about:blank';
        }, 500);
        break;

      case 'clearAppHistory':
        await window.go.main.App.ClearAppHistory(password);
        isConfirmModalOpen.set(false);
        showToast('Đã xóa lịch sử ứng dụng.', 'success');
        break;

      case 'clearWebHistory':
        await window.go.main.App.ClearWebHistory(password);
        isConfirmModalOpen.set(false);
        showToast('Đã xóa lịch sử duyệt web.', 'success');
        break;
    }
  } catch (error) {
    console.error(`Action ${action} failed:`, error);
    confirmModalError.set('Thao tác thất bại. Vui lòng kiểm tra lại mật khẩu.');
  }
}
