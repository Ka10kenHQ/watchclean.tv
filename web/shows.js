let shows = [];
let allShows = [];
let currentPage = 1;
let highlightedIndex = -1;
const pageSize = 50;

async function fetchShows() {
  try {
    const res = await fetch("/api/shows/images");
    if (!res.ok) throw new Error("Failed to fetch shows");
    shows = await res.json();
    allShows = [...shows];
    currentPage = 1;
    updateStatusBar();
    renderPage();
  } catch (err) {
    console.error("Error loading shows:", err);
  }
}

function updateStatusBar(status = "READY") {
  const statusBar = document.querySelector(".status-bar");
  const count = shows.length;
  statusBar.innerHTML = `
    <div class="status-item">
      <div class="status-dot"></div>
      <span>Shows: ${count}</span>
    </div>
    <div class="status-item">
      <span>Status: ${status}</span>
    </div>
  `;
}

function renderPage() {
  const start = (currentPage - 1) * pageSize;
  const end = start + pageSize;
  const currentItems = shows.slice(start, end);
  const grid = document.getElementById("movie-grid");
  grid.innerHTML = "";

  currentItems.forEach((show) => {
    const item = document.createElement("a");
    item.className = "movie-item";
    item.href = `/api/show/${show.id}`;

    const img = document.createElement("img");
    img.src = show.image || "/images/movie.jpg";
    img.alt = show.Title || show.title || show.id;
    img.className = "movie-poster";
    img.onerror = function() {
      this.src = "/images/movie.jpg";
    };

    const info = document.createElement("div");
    info.className = "movie-info";

    const title = document.createElement("div");
    title.className = "movie-title";
    title.textContent = show.Title || show.title || "Untitled";

    const englishTitle = document.createElement("div");
    englishTitle.className = "movie-english-title";
    englishTitle.textContent = show.TitleEnglish || show.titleEnglish || "";

    info.appendChild(title);
    if (englishTitle.textContent) {
      info.appendChild(englishTitle);
    }

    item.appendChild(img);
    item.appendChild(info);
    grid.appendChild(item);
  });

  const maxPage = Math.ceil(shows.length / pageSize);
  document.getElementById("page-number").textContent = `${String(currentPage).padStart(2, '0')} / ${String(maxPage || 1).padStart(2, '0')}`;
  updateNavigationButtons();
}

function updateNavigationButtons() {
  const maxPage = Math.ceil(shows.length / pageSize);
  document.getElementById("prev").disabled = currentPage === 1;
  document.getElementById("next").disabled = currentPage >= maxPage || maxPage === 0;
}

function showSearchResults(results, query) {
  const searchResults = document.getElementById("searchResults");
  if (!query || query.length < 2) {
    hideSearchResults();
    return;
  }
  if (results.length === 0) {
    searchResults.innerHTML = '<div class="no-results" style="padding: 1rem; color: var(--text-muted); font-size: 0.8rem; text-transform: uppercase;">No shows found</div>';
    searchResults.classList.add("show");
    return;
  }
  const limited = results.slice(0, 8);
  searchResults.innerHTML = limited
    .map(
      (show, idx) => `
<div class="search-result-item" data-index="${idx}" data-show-id="${show.id}">
  <img src="${show.image}" alt="${show.Title || show.title}" class="search-result-poster" />
  <div class="search-result-info">
    <div class="search-result-title">${show.Title || show.title}</div>
    <div class="search-result-id">${show.TitleEnglish || show.id}</div>
  </div>
</div>`,
    )
    .join("");
  searchResults.classList.add("show");
  highlightedIndex = -1;
}

function hideSearchResults() {
  const searchResults = document.getElementById("searchResults");
  if (searchResults) {
    searchResults.classList.remove("show");
  }
  highlightedIndex = -1;
}

function highlightResult(index) {
  const items = document.querySelectorAll(".search-result-item");
  items.forEach((item, i) => item.classList.toggle("highlighted", i === index));
  highlightedIndex = index;
}

function selectSearchResult(showId) {
  const show = allShows.find((m) => m.id === showId);
  if (show) {
    shows = [show];
    currentPage = 1;
    updateStatusBar("SELECTED");
    renderPage();
    document.getElementById("searchInput").value = show.Title || show.title;
    hideSearchResults();
  } else {
    // If not in local list, redirect to the show page
    window.location.href = `/api/show/${showId}`;
  }
}

async function performSearch(query) {
  if (!query || query.length < 2) {
    shows = [...allShows];
    currentPage = 1;
    updateStatusBar("READY");
    renderPage();
    hideSearchResults();
    return;
  }

  const localResults = allShows.filter((show) => {
    const title = (show.Title || show.title || "").toLowerCase();
    const english = (show.TitleEnglish || "").toLowerCase();
    const id = show.id.toLowerCase();
    const q = query.toLowerCase();
    return title.includes(q) || english.includes(q) || id.includes(q);
  });

  showSearchResults(localResults, query);

  try {
    const res = await fetch(`/api/search-shows?q=${encodeURIComponent(query)}`);
    if (res.ok) {
      let apiResults = await res.json();
      apiResults = apiResults.map((m) => ({
        id: m._id || m.id,
        Title: m.title || m.Title || "",
        TitleEnglish: m.titleEnglish || m.TitleEnglish || "",
        image: m.image || m.Image || "",
      }));
      
      // Merge with local results, avoiding duplicates
      const resultIds = new Set(localResults.map(m => m.id));
      const filteredApiResults = apiResults.filter(m => !resultIds.has(m.id));
      const combinedResults = [...localResults, ...filteredApiResults];
      
      showSearchResults(combinedResults, query);
    }
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

const debouncedSearch = debounce(performSearch, 200);

document.getElementById("prev").addEventListener("click", () => {
  if (currentPage > 1) {
    currentPage--;
    renderPage();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
});

document.getElementById("next").addEventListener("click", () => {
  if (currentPage < Math.ceil(shows.length / pageSize)) {
    currentPage++;
    renderPage();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
});

const searchInput = document.getElementById("searchInput");
if (searchInput) {
  searchInput.addEventListener("input", (e) => debouncedSearch(e.target.value));
  searchInput.addEventListener("keydown", (e) => {
    const items = document.querySelectorAll(".search-result-item");
    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        if (items.length) {
          highlightedIndex = Math.min(highlightedIndex + 1, items.length - 1);
          highlightResult(highlightedIndex);
        }
        break;
      case "ArrowUp":
        e.preventDefault();
        if (items.length) {
          highlightedIndex = Math.max(highlightedIndex - 1, -1);
          highlightResult(highlightedIndex);
        }
        break;
      case "Enter":
        e.preventDefault();
        if (highlightedIndex >= 0 && items[highlightedIndex]) {
          selectSearchResult(items[highlightedIndex].dataset.showId);
        }
        break;
      case "Escape":
        hideSearchResults();
        searchInput.blur();
        break;
    }
  });
}

const searchClear = document.getElementById("searchClear");
if (searchClear) {
  searchClear.addEventListener("click", () => {
    searchInput.value = "";
    performSearch("");
    searchClear.classList.remove("visible");
  });
  
  searchInput.addEventListener("input", () => {
    searchClear.classList.toggle("visible", searchInput.value.length > 0);
  });
}

document.addEventListener("click", (e) => {
  if (e.target.closest(".search-result-item")) {
    const showId = e.target.closest(".search-result-item").dataset.showId;
    selectSearchResult(showId);
  } else if (!e.target.closest(".search-bar")) {
    hideSearchResults();
  }
});

fetchShows();
