(function () {
  "use strict";

  const script = document.currentScript || document.querySelector("script[data-catalog-type]");
  const config = {
    type: script?.dataset.catalogType === "show" ? "show" : "movie",
    endpoint: script?.dataset.endpoint || "/api/movie-images",
    statusLabel: script?.dataset.statusLabel || "Movies",
    fallback: script?.dataset.fallback || "/images/movies.jpg",
  };

  const state = {
    items: [],
    allItems: [],
    currentPage: 1,
    highlightedIndex: -1,
    searchSeq: 0,
    searchCache: new Map(),
  };

  const pageSize = 50;
  const catalogCacheTTL = 10 * 60 * 1000;
  const refs = {
    grid: document.getElementById("movie-grid"),
    statusBar: document.querySelector(".status-bar"),
    pageNumber: document.getElementById("page-number"),
    pageNumberTop: document.getElementById("page-number-top"),
    pageNumbers: document.getElementById("page-numbers"),
    pageNumbersTop: document.getElementById("page-numbers-top"),
    prev: document.getElementById("prev"),
    next: document.getElementById("next"),
    prevTop: document.getElementById("prev-top"),
    nextTop: document.getElementById("next-top"),
    searchInput: document.getElementById("searchInput"),
    searchClear: document.getElementById("searchClear"),
    searchResults: document.getElementById("searchResults"),
  };

  function escapeHTML(value) {
    return String(value || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
  }

  function getID(item) {
    return item.id || item._id || item.ID || "";
  }

  function getTitle(item) {
    return item.Title || item.title || item.Name || item.name || "Untitled";
  }

  function getSubtitle(item, type) {
    return item.TitleEnglish || item.titleEnglish || item.Year || item.year || (type === "show" ? "TV show" : "Movie");
  }

  function detailURL(item, type) {
    const id = getID(item);
    const tmdb = item.tmdbId || item.TmdbID || item.tmdb_id || item.tmdb || "";
    if (id) return type === "show" ? `/api/show/${encodeURIComponent(String(id))}` : `/api/movie/${encodeURIComponent(String(id))}`;
    if (tmdb) return type === "show" ? `/shows/tmdb/${encodeURIComponent(String(tmdb))}` : `/movies/tmdb/${encodeURIComponent(String(tmdb))}`;
    return "";
  }

  function normalizeSearchItem(item, type) {
    const id = getID(item);
    const tmdb = item.tmdbId || item.TmdbID || item.tmdb_id || item.tmdb || "";
    const fallback = type === "show" ? "/images/movie.jpg" : "/images/movies.jpg";
    return {
      id: `${type}:${id || tmdb || getTitle(item)}`,
      type,
      title: getTitle(item),
      subtitle: getSubtitle(item, type),
      image: item.image || item.Image || fallback,
      url: detailURL(item, type),
      fallback,
    };
  }

  function updateStatusBar(status = "READY") {
    if (!refs.statusBar) return;
    refs.statusBar.innerHTML = `
      <div class="status-item">
        <div class="status-dot"></div>
        <span>${escapeHTML(config.statusLabel)}: ${state.items.length}</span>
      </div>
      <div class="status-item">
        <span>Status: ${escapeHTML(status)}</span>
      </div>`;
  }

  function cardHTML(item) {
    const fallback = escapeHTML(config.fallback);
    const image = escapeHTML((item.image && String(item.image).trim()) || config.fallback);
    const subtitle = getSubtitle(item, config.type);
    return `
<a class="movie-item" href="${escapeHTML(detailURL(item, config.type))}">
  <img src="${image}" alt="${escapeHTML(getTitle(item))}" class="movie-poster" loading="lazy" decoding="async" onerror="watchcleanPosterFallback(this,'${fallback}')" />
  <div class="movie-info">
    <div class="movie-title">${escapeHTML(getTitle(item))}</div>
    ${subtitle ? `<div class="movie-english-title">${escapeHTML(subtitle)}</div>` : ""}
  </div>
</a>`;
  }

  function renderPage() {
    if (!refs.grid) return;
    const start = (state.currentPage - 1) * pageSize;
    const currentItems = state.items.slice(start, start + pageSize);
    refs.grid.innerHTML = currentItems.length
      ? currentItems.map(cardHTML).join("")
      : '<div class="loading">No titles found.</div>';

    const maxPage = Math.ceil(state.items.length / pageSize) || 1;
    const pageText = `${String(state.currentPage).padStart(2, "0")} / ${String(maxPage).padStart(2, "0")}`;
    if (refs.pageNumber) refs.pageNumber.textContent = pageText;
    if (refs.pageNumberTop) refs.pageNumberTop.textContent = pageText;
    renderPageNumbers(maxPage);
    updateNavigationButtons(maxPage);
  }

  function pageNumbersHTML(maxPage) {
    const around = 2;
    const startPage = Math.max(1, state.currentPage - around);
    const endPage = Math.min(maxPage, state.currentPage + around);
    const parts = [];

    function btn(page) {
      parts.push(`<button class="page-number${page === state.currentPage ? " active" : ""}" data-page="${page}">${page}</button>`);
    }
    function dots() {
      parts.push('<span class="page-ellipsis">...</span>');
    }

    if (startPage > 1) {
      btn(1);
      if (startPage > 2) dots();
    }
    for (let page = startPage; page <= endPage; page++) btn(page);
    if (endPage < maxPage) {
      if (endPage < maxPage - 1) dots();
      btn(maxPage);
    }
    if (parts.length === 0) btn(1);
    return parts.join("");
  }

  function renderPageNumbers(maxPage) {
    const html = pageNumbersHTML(maxPage);
    if (refs.pageNumbers) refs.pageNumbers.innerHTML = html;
    if (refs.pageNumbersTop) refs.pageNumbersTop.innerHTML = html;
  }

  function updateNavigationButtons(maxPage) {
    const isFirst = state.currentPage === 1;
    const isLast = state.currentPage >= maxPage;
    [refs.prev, refs.prevTop].forEach((button) => {
      if (button) button.disabled = isFirst;
    });
    [refs.next, refs.nextTop].forEach((button) => {
      if (button) button.disabled = isLast;
    });
  }

  function goToPage(page) {
    const maxPage = Math.ceil(state.items.length / pageSize) || 1;
    state.currentPage = Math.min(Math.max(page, 1), maxPage);
    renderPage();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function showSearchResults(results, query) {
    if (!refs.searchResults) return;
    if (!query || query.length < 2) {
      hideSearchResults();
      return;
    }

    const limited = results.filter((item) => item.url).slice(0, 8);
    if (limited.length === 0) {
      refs.searchResults.innerHTML = '<div class="no-results">No movies or shows found</div>';
      refs.searchResults.classList.add("show");
      return;
    }

    refs.searchResults.innerHTML = limited
      .map((item, index) => {
        const label = item.type === "show" ? "TV" : "Movie";
        return `
<div class="search-result-item" data-index="${index}" data-url="${escapeHTML(item.url)}">
  <img src="${escapeHTML(item.image)}" alt="" class="search-result-poster" loading="lazy" decoding="async" onerror="watchcleanPosterFallback(this,'${escapeHTML(item.fallback)}')" />
  <div class="search-result-info">
    <div class="search-result-title">${escapeHTML(item.title)}</div>
    <div class="search-result-id">${label} · ${escapeHTML(item.subtitle)}</div>
  </div>
</div>`;
      })
      .join("");
    refs.searchResults.classList.add("show");
    state.highlightedIndex = -1;
  }

  function hideSearchResults() {
    refs.searchResults?.classList.remove("show");
    state.highlightedIndex = -1;
  }

  function highlightResult(index) {
    const items = document.querySelectorAll(".search-result-item");
    items.forEach((item, itemIndex) => item.classList.toggle("highlighted", itemIndex === index));
    state.highlightedIndex = index;
  }

  async function queryBothCatalogs(query) {
    const key = query.trim().toLowerCase();
    if (state.searchCache.has(key)) return state.searchCache.get(key);

    const [movieRes, showRes] = await Promise.all([
      fetch(`/api/search?q=${encodeURIComponent(query)}`),
      fetch(`/api/shows/search?q=${encodeURIComponent(query)}`),
    ]);
    const [movies, shows] = await Promise.all([
      movieRes.ok ? movieRes.json() : [],
      showRes.ok ? showRes.json() : [],
    ]);
    const results = [
      ...(Array.isArray(movies) ? movies.map((movie) => normalizeSearchItem(movie, "movie")) : []),
      ...(Array.isArray(shows) ? shows.map((show) => normalizeSearchItem(show, "show")) : []),
    ];
    state.searchCache.set(key, results);
    return results;
  }

  async function performSearch(query) {
    const q = query.trim();
    if (q.length < 2) {
      state.searchSeq++;
      state.items = [...state.allItems];
      state.currentPage = 1;
      updateStatusBar("READY");
      renderPage();
      hideSearchResults();
      return;
    }

    const mySeq = ++state.searchSeq;
    const localResults = state.allItems
      .filter((item) => {
        const needle = q.toLowerCase();
        return (
          getTitle(item).toLowerCase().includes(needle) ||
          getSubtitle(item, config.type).toLowerCase().includes(needle) ||
          String(getID(item)).toLowerCase().includes(needle)
        );
      })
      .map((item) => normalizeSearchItem(item, config.type));

    showSearchResults(localResults, q);

    try {
      const apiResults = await queryBothCatalogs(q);
      if (mySeq !== state.searchSeq) return;
      const seen = new Set(localResults.map((item) => item.id));
      showSearchResults([...localResults, ...apiResults.filter((item) => !seen.has(item.id))], q);
    } catch (err) {
      console.error("API search error:", err);
    }
  }

  function debounce(fn, delay) {
    let timeout;
    return (...args) => {
      clearTimeout(timeout);
      timeout = setTimeout(() => fn(...args), delay);
    };
  }

  async function fetchJSONWithCache(url, ttl) {
    const key = `watchlean:${url}`;
    try {
      const cached = JSON.parse(localStorage.getItem(key) || "null");
      if (cached && Date.now() - cached.savedAt < ttl) {
        return cached.data;
      }
    } catch (err) {
      localStorage.removeItem(key);
    }

    const res = await fetch(url);
    if (!res.ok) throw new Error(`Failed to fetch ${url}`);
    const data = await res.json();
    try {
      localStorage.setItem(key, JSON.stringify({ savedAt: Date.now(), data }));
    } catch (err) {
      // Ignore quota/private-mode failures; network data still renders.
    }
    return data;
  }

  async function loadCatalog() {
    try {
      state.items = await fetchJSONWithCache(config.endpoint, catalogCacheTTL);
      state.allItems = [...state.items];
      state.currentPage = 1;
      updateStatusBar();
      renderPage();
    } catch (err) {
      console.error(`Error loading ${config.statusLabel}:`, err);
      if (refs.grid) refs.grid.innerHTML = `<div class="loading">Could not load ${escapeHTML(config.statusLabel)}.</div>`;
    }
  }

  const debouncedSearch = debounce(performSearch, 200);

  document.addEventListener("click", (event) => {
    const pageButton = event.target.closest(".page-number[data-page]");
    if (pageButton) {
      goToPage(Number(pageButton.dataset.page));
      return;
    }
    const result = event.target.closest(".search-result-item");
    if (result) {
      const url = result.dataset.url;
      if (url) window.location.href = url;
      return;
    }
    if (!event.target.closest(".search-container")) hideSearchResults();
  });

  refs.prev?.addEventListener("click", () => goToPage(state.currentPage - 1));
  refs.next?.addEventListener("click", () => goToPage(state.currentPage + 1));
  refs.prevTop?.addEventListener("click", () => goToPage(state.currentPage - 1));
  refs.nextTop?.addEventListener("click", () => goToPage(state.currentPage + 1));

  refs.searchInput?.addEventListener("input", (event) => {
    refs.searchClear?.classList.toggle("visible", event.target.value.length > 0);
    debouncedSearch(event.target.value);
  });

  refs.searchInput?.addEventListener("keydown", (event) => {
    const items = document.querySelectorAll(".search-result-item");
    if (event.key === "ArrowDown") {
      event.preventDefault();
      if (items.length) highlightResult(Math.min(state.highlightedIndex + 1, items.length - 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      if (items.length) highlightResult(Math.max(state.highlightedIndex - 1, -1));
    } else if (event.key === "Enter") {
      event.preventDefault();
      if (state.highlightedIndex >= 0 && items[state.highlightedIndex]) {
        const url = items[state.highlightedIndex].dataset.url;
        if (url) window.location.href = url;
      }
    } else if (event.key === "Escape") {
      hideSearchResults();
      refs.searchInput.blur();
    }
  });

  refs.searchClear?.addEventListener("click", () => {
    refs.searchInput.value = "";
    refs.searchClear.classList.remove("visible");
    performSearch("");
    refs.searchInput.focus();
  });

  loadCatalog();
})();
