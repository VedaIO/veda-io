document.addEventListener('DOMContentLoaded', () => {
  const statusElement = document.getElementById('status');
  const blockSiteBtn = document.getElementById('block-site-btn');

  // For now, just show a static status.
  statusElement.textContent = 'Sẵn sàng';

  blockSiteBtn.addEventListener('click', () => {
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      if (tabs[0] && tabs[0].url) {
        try {
          const url = new URL(tabs[0].url);
          const domain = url.hostname;
          // Send a message to the background script to block the domain.
          chrome.runtime.sendMessage({ type: 'add_to_web_blocklist', payload: domain }, (response) => {
            // You can handle the response here, e.g., close the popup or show a confirmation.
            window.close();
          });
        } catch (e) {
          // Handle invalid URLs
          
        }
      }
    });
  });
});
