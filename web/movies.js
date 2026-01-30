let movies = [];
let allMovies = [];
let currentPage = 1;
let highlightedIndex = -1;
const pageSize = 50;

async function fetchMovies() {
  try {
    const res = await fetch("/api/movie-images");
    if (!res.ok) throw new Error("Failed to fetch movies");
    movies = await res.json();
    allMovies = [...movies];
    currentPage = 1;
    updateStatusBar();
    renderPage();
  } catch (err) {
    console.error("Error loading movies:", err);
  }
}

function updateStatusBar(status = "READY") {
  const statusBar = document.querySelector(".status-bar");
  const count = movies.length;
  statusBar.innerHTML = `
    <div class="status-item">
      <div class="status-dot"></div>
      <span>Movies: ${count}</span>
    </div>
    <div class="status-item">
      <span>Status: ${status}</span>
    </div>
  `;
}

function renderPage() {
  const start = (currentPage - 1) * pageSize;
  const end = start + pageSize;
  const currentItems = movies.slice(start, end);
  const grid = document.getElementById("movie-grid");
  grid.innerHTML = "";

  currentItems.forEach((movie) => {
    const item = document.createElement("a");
    item.className = "movie-item";
    item.href = `/api/movie/${movie.id}`;

    const img = document.createElement("img");
    img.src = movie.image || "/images/movies.jpg";
    img.alt = movie.Title || movie.title || movie.id;
    img.className = "movie-poster";
    img.onerror = function() {
      this.src = "/images/movies.jpg";
    };

    const info = document.createElement("div");
    info.className = "movie-info";

    const title = document.createElement("div");
    title.className = "movie-title";
    title.textContent = movie.Title || movie.title || "Untitled";

    const englishTitle = document.createElement("div");
    englishTitle.className = "movie-english-title";
    englishTitle.textContent = movie.TitleEnglish || movie.titleEnglish || "";

    info.appendChild(title);
    if (englishTitle.textContent) {
      info.appendChild(englishTitle);
    }

    item.appendChild(img);
    item.appendChild(info);
    grid.appendChild(item);
  });

  const maxPage = Math.ceil(movies.length / pageSize);
  const pageText = `${String(currentPage).padStart(2, '0')} / ${String(maxPage || 1).padStart(2, '0')}`;
  document.getElementById("page-number").textContent = pageText;
  document.getElementById("page-number-top").textContent = pageText;
  renderPageNumbers();
  updateNavigationButtons();
}

function renderPageNumbers() {
  const maxPage = Math.ceil(movies.length / pageSize) || 1;
  const pageNumbersContainer = document.getElementById("page-numbers");
  const pageNumbersContainerTop = document.getElementById("page-numbers-top");
  
  if (!pageNumbersContainer || !pageNumbersContainerTop) {
    console.error('Pagination containers not found!');
    return;
  }
  
  pageNumbersContainer.innerHTML = "";
  pageNumbersContainerTop.innerHTML = "";

  // Show at least the current page button even if there's only 1 page
  if (maxPage <= 1) {
    addPageNumber(1, pageNumbersContainer);
    addPageNumber(1, pageNumbersContainerTop);
    return;
  }

  const showAround = 2;
  let startPage = Math.max(1, currentPage - showAround);
  let endPage = Math.min(maxPage, currentPage + showAround);

  // Render bottom pagination
  if (startPage > 1) {
    addPageNumber(1, pageNumbersContainer);
    if (startPage > 2) {
      addEllipsis(pageNumbersContainer);
    }
  }

  for (let i = startPage; i <= endPage; i++) {
    addPageNumber(i, pageNumbersContainer);
  }

  if (endPage < maxPage) {
    if (endPage < maxPage - 1) {
      addEllipsis(pageNumbersContainer);
    }
    addPageNumber(maxPage, pageNumbersContainer);
  }

  // Render top pagination (same logic)
  if (startPage > 1) {
    addPageNumber(1, pageNumbersContainerTop);
    if (startPage > 2) {
      addEllipsis(pageNumbersContainerTop);
    }
  }

  for (let i = startPage; i <= endPage; i++) {
    addPageNumber(i, pageNumbersContainerTop);
  }

  if (endPage < maxPage) {
    if (endPage < maxPage - 1) {
      addEllipsis(pageNumbersContainerTop);
    }
    addPageNumber(maxPage, pageNumbersContainerTop);
  }
}

function addPageNumber(page, container) {
  const button = document.createElement("button");
  button.className = "page-number";
  if (page === currentPage) {
    button.classList.add("active");
  }
  button.textContent = page;
  button.onclick = () => {
    currentPage = page;
    renderPage();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };
  container.appendChild(button);
}

function addEllipsis(container) {
  const span = document.createElement("span");
  span.className = "page-ellipsis";
  span.textContent = "...";
  container.appendChild(span);
}

function updateNavigationButtons() {
  const maxPage = Math.ceil(movies.length / pageSize);
  const isFirstPage = currentPage === 1;
  const isLastPage = currentPage >= maxPage || maxPage === 0;
  
  document.getElementById("prev").disabled = isFirstPage;
  document.getElementById("next").disabled = isLastPage;
  document.getElementById("prev-top").disabled = isFirstPage;
  document.getElementById("next-top").disabled = isLastPage;
}

function showSearchResults(results, query) {
  const searchResults = document.getElementById("searchResults");
  if (!query || query.length < 2) {
    hideSearchResults();
    return;
  }
  if (results.length === 0) {
    searchResults.innerHTML = '<div class="no-results" style="padding: 1rem; color: var(--text-muted); font-size: 0.8rem; text-transform: uppercase;">No movies found</div>';
    searchResults.classList.add("show");
    return;
  }
  const limited = results.slice(0, 8);
  searchResults.innerHTML = limited
    .map(
      (movie, idx) => `
<div class="search-result-item" data-index="${idx}" data-movie-id="${movie.id}">
  <img src="${movie.image}" alt="${movie.Title || movie.title}" class="search-result-poster" />
  <div class="search-result-info">
    <div class="search-result-title">${movie.Title || movie.title}</div>
    <div class="search-result-id">${movie.TitleEnglish || movie.id}</div>
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

function selectSearchResult(movieId) {
  const movie = allMovies.find((m) => m.id === movieId);
  if (movie) {
    movies = [movie];
    currentPage = 1;
    updateStatusBar("SELECTED");
    renderPage();
    document.getElementById("searchInput").value = movie.Title || movie.title;
    hideSearchResults();
  } else {
    // If not in local list, redirect to the movie page
    window.location.href = `/api/movie/${movieId}`;
  }
}

async function performSearch(query) {
  if (!query || query.length < 2) {
    movies = [...allMovies];
    currentPage = 1;
    updateStatusBar("READY");
    renderPage();
    hideSearchResults();
    return;
  }

  const localResults = allMovies.filter((movie) => {
    const title = (movie.Title || movie.title || "").toLowerCase();
    const english = (movie.TitleEnglish || "").toLowerCase();
    const id = movie.id.toLowerCase();
    const q = query.toLowerCase();
    return title.includes(q) || english.includes(q) || id.includes(q);
  });

  showSearchResults(localResults, query);

  try {
    const res = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
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
  if (currentPage < Math.ceil(movies.length / pageSize)) {
    currentPage++;
    renderPage();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
});

document.getElementById("prev-top").addEventListener("click", () => {
  if (currentPage > 1) {
    currentPage--;
    renderPage();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
});

document.getElementById("next-top").addEventListener("click", () => {
  if (currentPage < Math.ceil(movies.length / pageSize)) {
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
          selectSearchResult(items[highlightedIndex].dataset.movieId);
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
    const movieId = e.target.closest(".search-result-item").dataset.movieId;
    selectSearchResult(movieId);
  } else if (!e.target.closest(".search-bar")) {
    hideSearchResults();
  }
});

fetchMovies();
