(function () {
  "use strict";

  const SAMPLE_SIZE = 100;
  const CATALOG_CACHE_TTL = 10 * 60 * 1000;
  const state = {
    all: [],
    visible: [],
    filter: "all",
    query: "",
  };

  const grid = document.getElementById("landingGrid");
  const searchInput = document.getElementById("landingSearch");
  const filterButtons = Array.from(document.querySelectorAll(".landing-filter"));
  const shuffleButton = document.getElementById("shuffleLanding");

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

  function normalize(item, type) {
    const id = getID(item);
    const title = item.Title || item.title || "Untitled";
    const subtitle = item.TitleEnglish || item.titleEnglish || (type === "show" ? "TV show" : "Movie");
    const fallback = type === "show" ? "/images/movie.jpg" : "/images/movies.jpg";
    return {
      id: `${type}:${id || title}`,
      rawID: id,
      type,
      title,
      subtitle,
      image: item.image || item.Image || fallback,
      fallback,
      url: type === "show" ? `/api/show/${encodeURIComponent(String(id))}` : `/api/movie/${encodeURIComponent(String(id))}`,
    };
  }

  function shuffle(items) {
    const copy = [...items];
    for (let i = copy.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [copy[i], copy[j]] = [copy[j], copy[i]];
    }
    return copy;
  }

  function pickRandomShelf() {
    state.visible = shuffle(state.all).slice(0, SAMPLE_SIZE);
  }

  function getFilteredItems() {
    const query = state.query.trim().toLowerCase();
    let items = query ? state.all : state.visible;

    if (state.filter !== "all") {
      items = items.filter((item) => item.type === state.filter);
    }

    if (query) {
      items = items.filter((item) => {
        return (
          item.title.toLowerCase().includes(query) ||
          item.subtitle.toLowerCase().includes(query) ||
          String(item.rawID).toLowerCase().includes(query)
        );
      });
      items = items.slice(0, SAMPLE_SIZE);
    }

    return items;
  }

  function render() {
    if (!grid) return;
    const items = getFilteredItems().filter((item) => item.rawID);

    if (items.length === 0) {
      grid.innerHTML = '<div class="loading">No matches found.</div>';
      return;
    }

    grid.innerHTML = items
      .map((item) => {
        const label = item.type === "show" ? "TV show" : "Movie";
        const image = escapeHTML(item.image);
        const fallback = escapeHTML(item.fallback);
        return `
<a class="landing-card" href="${escapeHTML(item.url)}">
  <span class="landing-card-type">${label}</span>
  <img src="${image}" alt="" class="landing-card-poster" loading="lazy" decoding="async" onerror="watchcleanPosterFallback(this,'${fallback}')" />
  <div class="landing-card-body">
    <h3>${escapeHTML(item.title)}</h3>
    <p>${escapeHTML(item.subtitle)}</p>
  </div>
</a>`;
      })
      .join("");
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

  async function load() {
    try {
      const [movies, shows] = await Promise.all([
        fetchJSONWithCache("/api/movie-images", CATALOG_CACHE_TTL),
        fetchJSONWithCache("/api/shows/images", CATALOG_CACHE_TTL),
      ]);

      state.all = [
        ...(Array.isArray(movies) ? movies.map((movie) => normalize(movie, "movie")) : []),
        ...(Array.isArray(shows) ? shows.map((show) => normalize(show, "show")) : []),
      ];

      pickRandomShelf();
      render();
    } catch (err) {
      console.error("Landing load error:", err);
      if (grid) grid.innerHTML = '<div class="loading">Could not load the vault.</div>';
    }
  }

  searchInput?.addEventListener(
    "input",
    debounce((event) => {
      state.query = event.target.value;
      render();
    }, 120),
  );

  filterButtons.forEach((button) => {
    button.addEventListener("click", () => {
      state.filter = button.dataset.filter || "all";
      filterButtons.forEach((item) => item.classList.toggle("is-active", item === button));
      render();
    });
  });

  shuffleButton?.addEventListener("click", () => {
    state.query = "";
    if (searchInput) searchInput.value = "";
    pickRandomShelf();
    render();
  });

  load();
})();
