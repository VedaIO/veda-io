import { writable } from 'svelte/store';

export const isExtensionInstalled = writable<boolean | null>(null);

export function checkExtension(): void {
  // If we already have a definitive answer, don't re-run.
  let installed = null;
  isExtensionInstalled.subscribe((val) => (installed = val))();
  if (installed !== null) {
    return;
  }

  const observer = new MutationObserver((mutations, obs) => {
    const idDiv = document.getElementById('procguard-extension-id');
    if (idDiv && idDiv.textContent) {
      const extensionId = idDiv.textContent;

      try {
        chrome.runtime.sendMessage(
          extensionId,
          { message: 'is_installed' },
          (response) => {
            if (chrome.runtime.lastError) {
              isExtensionInstalled.set(false);
            } else {
              if (response && response.status === 'installed') {
                isExtensionInstalled.set(true);
                // Register the extension with the backend
                window.go.main.App.RegisterExtension(extensionId).catch((err) =>
                  console.error('Failed to register extension:', err)
                );
              }
            }
          }
        );
      } catch {
        isExtensionInstalled.set(false);
      }
      obs.disconnect();
      return;
    }
  });

  observer.observe(document.body, {
    childList: true,
    subtree: true,
  });

  // Stop observing after a timeout if the div is not found.
  setTimeout(() => {
    observer.disconnect();
    let installed_timeout = null;
    isExtensionInstalled.subscribe((val) => (installed_timeout = val))();
    if (installed_timeout === null) {
      isExtensionInstalled.set(false);
    }
  }, 1000); // Shorten timeout to 1 second
}
