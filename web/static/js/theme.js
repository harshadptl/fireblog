// Loaded synchronously in <head> so a saved light/dark choice applies before
// the first paint (no flash). Without a saved choice the OS setting is used.
(function () {
  try {
    var t = localStorage.getItem('theme');
    if (t === 'light' || t === 'dark') document.documentElement.dataset.theme = t;
  } catch (e) { /* storage unavailable: follow the OS */ }
})();
