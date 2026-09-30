# Browser extensions

The popup reads only the active tab URL after a user click. It does not inspect page content or run a background content script.

Build both unpacked extension folders with:

```bash
cd extensions
npm run build
```

Load `dist/chrome` or `dist/firefox` as an unpacked/temporary extension. The local Angular workspace URL is currently a constant in `shared/src/popup.js`; replace it with the deployed frontend URL during release packaging.

