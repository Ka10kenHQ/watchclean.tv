/**
 * Safe poster/image fallback: one swap to app fallback, then a tiny SVG data-URL
 * that always loads (stops infinite onerror loops / flicker).
 */
(function () {
  'use strict';

  var PLACEHOLDER_SVG =
    'data:image/svg+xml,' +
    encodeURIComponent(
      '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 300" width="200" height="300">' +
        '<rect fill="%23e8ecf1" width="200" height="300"/>' +
        '<text x="100" y="155" text-anchor="middle" fill="%2394a3b8" font-size="11" font-family="system-ui,sans-serif">No image</text>' +
        '</svg>'
    );

  /**
   * @param {HTMLImageElement} img
   * @param {string} fallbackSrc site path, e.g. /images/movies.jpg
   */
  function watchcleanPosterFallback(img, fallbackSrc) {
    if (!img || !fallbackSrc) return;
    if (img.dataset.watchcleanPosterDone === '1') return;
    if (img.src.indexOf('data:image/svg+xml') === 0) return;

    var absFb;
    try {
      absFb = new URL(fallbackSrc, window.location.origin).href;
    } catch (e) {
      absFb = fallbackSrc;
    }

    if (img.src !== absFb) {
      img.src = fallbackSrc;
      return;
    }

    img.dataset.watchcleanPosterDone = '1';
    img.classList.add('poster--placeholder');
    img.src = PLACEHOLDER_SVG;
  }

  function attachPosterFallback(img, fallbackSrc) {
    if (!img) return;
    img.loading = 'lazy';
    img.decoding = 'async';
    img.addEventListener(
      'error',
      function () {
        watchcleanPosterFallback(img, fallbackSrc);
      },
      { passive: true }
    );
  }

  window.watchcleanPosterFallback = watchcleanPosterFallback;
  window.watchcleanAttachPosterFallback = attachPosterFallback;
})();
