const workspaceUrl = 'http://localhost:4200/';
const captureButton = document.querySelector('#capture');
const status = document.querySelector('#status');

captureButton.addEventListener('click', async () => {
  try {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    if (!tab?.url || !/^https?:/.test(tab.url)) {
      throw new Error('Open a public job posting first.');
    }
    const destination = new URL(workspaceUrl);
    destination.searchParams.set('jobUrl', tab.url);
    await chrome.tabs.create({ url: destination.toString() });
    window.close();
  } catch (error) {
    status.textContent = error instanceof Error ? error.message : 'Could not capture this tab.';
  }
});

