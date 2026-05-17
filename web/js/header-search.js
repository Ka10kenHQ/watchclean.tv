(function () {
  "use strict";

  const input = document.getElementById("searchInput");
  const clear = document.getElementById("searchClear");
  const results = document.getElementById("searchResults");
  if (!input || !results) return;

  let highlightedIndex = -1;
  let searchSeq = 0;

  function escapeHTML(value) {
    return String(value || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
  }

  function getField(item, names) {
    for (const name of names) {
      if (item && item[name] !== undefined && item[name] !== null && item[name] !== "") {
        return item[name];
      }
    }
    return "";
  }

  function getURL(item, type) {
    const id = getField(item, ["id", "_id", "ID"]);
    if (id) {
      const idBase = type === "show" ? "/api/show/" : "/api/movie/";
      return `${idBase}${encodeURIComponent(String(id))}`;
    }
    const tmdb = getField(item, ["tmdbId", "TmdbID", "tmdb_id", "tmdb"]);
    if (tmdb) {
      const tmdbBase = type === "show" ? "/shows/tmdb/" : "/movies/tmdb/";
      return `${tmdbBase}${encodeURIComponent(String(tmdb))}`;
    }
    return "";
  }

  function normalize(item, type) {
    const fallbackPoster = type === "show" ? "/images/movie.jpg" : "/images/movies.jpg";
    return {
      type,
      title: getField(item, ["title", "Title", "name", "Name"]) || "Untitled",
      subtitle: getField(item, ["titleEnglish", "TitleEnglish", "year", "Year", "tmdbId", "TmdbID"]) || (type === "show" ? "TV show" : "Movie"),
      image: getField(item, ["image", "Image"]) || fallbackPoster,
      fallbackPoster,
      url: getURL(item, type),
    };
  }

  function hideResults() {
    results.classList.remove("show");
    results.innerHTML = "";
    highlightedIndex = -1;
  }

  function highlight(index) {
    const items = results.querySelectorAll(".search-result-item");
    items.forEach((item, i) => item.classList.toggle("highlighted", i === index));
    highlightedIndex = index;
  }

  function render(items, query) {
    if (!query || query.length < 2) {
      hideResults();
      return;
    }

    const normalized = items.filter((item) => item.url).slice(0, 8);
    if (normalized.length === 0) {
      results.innerHTML = '<div class="no-results">No movies or shows found</div>';
      results.classList.add("show");
      return;
    }

    results.innerHTML = normalized
      .map((item, idx) => {
        const title = escapeHTML(item.title);
        const label = item.type === "show" ? "TV" : "Movie";
        const subtitle = escapeHTML(item.subtitle);
        const image = escapeHTML(item.image);
        const url = escapeHTML(item.url);
        const fallbackPoster = escapeHTML(item.fallbackPoster);
        return `
<a class="search-result-item" data-index="${idx}" href="${url}">
  <img src="${image}" alt="" class="search-result-poster" loading="lazy" decoding="async" onerror="watchcleanPosterFallback(this,'${fallbackPoster}')" />
  <div class="search-result-info">
    <div class="search-result-title">${title}</div>
    <div class="search-result-id">${label} · ${subtitle}</div>
  </div>
</a>`;
      })
      .join("");
    results.classList.add("show");
    highlightedIndex = -1;
  }

  function debounce(fn, delay) {
    let timeout;
    return (...args) => {
      clearTimeout(timeout);
      timeout = setTimeout(() => fn(...args), delay);
    };
  }

  async function search(query) {
    const q = query.trim();
    if (q.length < 2) {
      searchSeq++;
      hideResults();
      return;
    }

    const mySeq = ++searchSeq;
    try {
      const [movieRes, showRes] = await Promise.all([
        fetch(`/api/search?q=${encodeURIComponent(q)}`),
        fetch(`/api/shows/search?q=${encodeURIComponent(q)}`),
      ]);
      if (mySeq !== searchSeq) return;
      const [movies, shows] = await Promise.all([
        movieRes.ok ? movieRes.json() : [],
        showRes.ok ? showRes.json() : [],
      ]);
      if (mySeq !== searchSeq) return;
      render(
        [
          ...(Array.isArray(movies) ? movies.map((movie) => normalize(movie, "movie")) : []),
          ...(Array.isArray(shows) ? shows.map((show) => normalize(show, "show")) : []),
        ],
        q,
      );
    } catch (err) {
      console.error("Header search error:", err);
    }
  }

  const debouncedSearch = debounce(search, 220);

  input.addEventListener("input", (event) => {
    const value = event.target.value;
    clear?.classList.toggle("visible", value.length > 0);
    debouncedSearch(value);
  });

  input.addEventListener("keydown", (event) => {
    const items = results.querySelectorAll(".search-result-item");
    if (event.key === "ArrowDown") {
      event.preventDefault();
      if (items.length) highlight(Math.min(highlightedIndex + 1, items.length - 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      if (items.length) highlight(Math.max(highlightedIndex - 1, -1));
    } else if (event.key === "Enter") {
      if (highlightedIndex >= 0 && items[highlightedIndex]) {
        event.preventDefault();
        window.location.href = items[highlightedIndex].href;
      } else if (input.value.trim()) {
        event.preventDefault();
        window.location.href = `/movies?q=${encodeURIComponent(input.value.trim())}`;
      }
    } else if (event.key === "Escape") {
      hideResults();
      input.blur();
    }
  });

  clear?.addEventListener("click", () => {
    input.value = "";
    clear.classList.remove("visible");
    hideResults();
    input.focus();
  });

  document.addEventListener("click", (event) => {
    if (!event.target.closest(".search-container")) {
      hideResults();
    }
  });
})();
