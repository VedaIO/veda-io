import { writable } from 'svelte/store';

export const isUninstallModalOpen = writable(false);
export const uninstallPassword = writable('');
export const uninstallError = writable('');

export function openUninstallModal() {
  uninstallPassword.set('');
  uninstallError.set('');
  isUninstallModalOpen.set(true);
}

export async function handleUninstallSubmit() {
  let password = '';
  uninstallPassword.subscribe((value) => (password = value))();

  try {
    await window.go.main.App.Uninstall(password);
    isUninstallModalOpen.set(false);
    // Give the modal a moment to close before closing the page
    setTimeout(() => {
      window.location.href = 'about:blank';
    }, 500);
  } catch (error) {
    console.error('Uninstall error:', error);
    uninstallError.set('Gỡ cài đặt thất bại. Vui lòng kiểm tra lại mật khẩu.');
  }
}
