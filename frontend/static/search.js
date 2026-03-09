const search = document.getElementById("search");
const nav = document.querySelectorAll("body > header > nav")[0];
const queryInput = document.getElementById("query");
const dialog = document.getElementById("dialog");
const results = document.getElementById("results");
let pagination = document.getElementById("pagination");
let facets = document.querySelector(".facets");
let titleParts = document.title.split(/ [\p{Dash}] /u);
let siteTitle = titleParts.slice(titleParts.length > 1 ? 1 : 0).join(" ‐ ");

const resultsRange = new Range();
resultsRange.setStart(results, 0);
resultsRange.setEnd(results, results.childNodes.length);

const detailsRange = new Range();
detailsRange.setStartAfter(dialog.firstElementChild);
detailsRange.setEndAfter(dialog.lastElementChild);

let closedByPopstate = false;

const pageHistory = {
  state: null,

  init() {
    const url = new URL(location);
    this.state = history.state || {
      url: url.toString(),
      title: document.title,
      input: url.searchParams.get("query"),
      results: resultsRange.cloneContents().innerHTML || null,
    };
  },

  push(url) {
    history.pushState(this.state, null, url);
  },

  replace(url) {
    history.replaceState(this.state, null, url);
  },

  pushDetail(url) {
    history.pushState({ detail: true, searchState: this.state }, null, url);
  },

  restore(historyState) {
    if (historyState != null) {
      this.state = historyState.searchState || historyState;
    }
  },
};

pageHistory.init();

let escapePolicy = null;
if (window.trustedTypes && trustedTypes.createPolicy) {
  escapePolicy = trustedTypes.createPolicy("fetch", {
    createHTML: (string) => string,
  });
}

/**
 *
 * @param {MouseEvent} ev
 */
function openSiblingDialog(ev) {
  if (!(ev.shiftKey || ev.altKey || ev.ctrlKey || ev.metaKey)) {
    const dialog = ev.target.nextElementSibling;
    dialog.showModal();
    dialog.querySelector("button").addEventListener("click", function () {
      dialog.close();
    });
  }
}

/**
 *
 * @param {HTMLElement} results
 */
function addOpenDialogListeners(results) {
  results.querySelectorAll("a.open-dialog").forEach(function (element) {
    element.addEventListener("click", handleDialogOpen);
  });
  results.querySelectorAll("a.open-sibling-dialog").forEach(function (element) {
    element.addEventListener("click", openSiblingDialog);
  });
}

/**
 *
 * @param {MouseEvent} ev
 */
function updateSearch(ev) {
  const url = new URL(ev.target.href);
  getResults(url);
  ev.preventDefault();
}

/**
 *
 * @param {HTMLElement} parent
 */
function addUpdateSearchListeners(parent) {
  parent
    .querySelectorAll("a")
    .forEach((child) => child.addEventListener("click", updateSearch));
}

/**
 *
 * @param {string|null} title
 * @param {boolean} updateState
 */
function updateTitle(title, updateState) {
  if (title) {
    title = title.replaceAll("&hyphen;", "\u2010");
    document.title = title + " " + siteTitle;
    if (updateState) {
      pageHistory.state.title = title;
    }
  } else {
    document.title = pageHistory.state.title;
  }
}

/**
 * @param {string} html
 */
function renderResults(html) {
  const fragment = resultsRange.createContextualFragment(
    escapePolicy !== null ? escapePolicy.createHTML(html) : html,
  );
  resultsRange.deleteContents();
  resultsRange.insertNode(fragment);
  addOpenDialogListeners(results);

  pagination = results.querySelector("#pagination");
  if (pagination !== null) {
    addUpdateSearchListeners(pagination);
  }

  facets = results.querySelector(".facets");
  if (facets !== null) {
    addUpdateSearchListeners(facets);
  }
}

/**
 *
 * @param {URL} url
 */
async function getResults(url) {
  try {
    pageHistory.state.url = url.toJSON();
    pageHistory.push(url);
    const res = await fetch(url, {
      headers: {
        fetch: "true",
      },
    });
    updateTitle(res.headers.get("title"), true);

    // render errors sent as HTML as well as OK responses
    if (res.headers.get("content-type").startsWith("text/html")) {
      pageHistory.state.results = await res.text();
      pageHistory.replace(url);
      renderResults(pageHistory.state.results);
    } else {
      throw new Error(`${res.status} ${res.statusText}: ${await res.text()}`);
    }
  } catch (error) {
    resultsRange.deleteContents();
    resultsRange.insertNode(new Text(error.message));
    console.error("fetch failed", error);
  }
}

queryInput.addEventListener("input", function (ev) {
  for (const el of nav.children) {
    if (el.nodeName === "A") {
      const url = new URL(el.href);
      if (ev.target.value && !el.classList.contains("current")) {
        url.searchParams.set("query", ev.target.value);
      } else {
        url.searchParams.delete("query");
      }
      el.href = url.toString();
    }
  }
});

search.addEventListener("submit", function (ev) {
  const url = new URL(this.action);
  const formData = new FormData(this);
  url.search = new URLSearchParams(formData).toString();
  pageHistory.state.input = formData.get("query");
  getResults(url);
  ev.preventDefault();
});

if (results !== null) {
  addOpenDialogListeners(results);
  if (pagination !== null) {
    addUpdateSearchListeners(pagination);
  }
  if (facets !== null) {
    addUpdateSearchListeners(facets);
  }
}

document.querySelector("a.current").addEventListener("click", function (ev) {
  search.reset();
  pageHistory.state.input = null;
  resultsRange.deleteContents();
  pageHistory.state.results = "";
  pageHistory.push(ev.target.href);
  ev.preventDefault();
  nav.querySelectorAll("a:not(.current)").forEach(function (node) {
    const u = new URL(node.href);
    u.search = "";
    node.href = u.toString();
  });
  queryInput.value = "";
});

/**
 *
 * @param {string} html
 */
function renderDetails(html) {
  const fragment = detailsRange.createContextualFragment(
    escapePolicy !== null ? escapePolicy.createHTML(html) : html,
  );
  detailsRange.insertNode(fragment);
  dialog.showModal();
}

dialog.addEventListener("close", function () {
  detailsRange.deleteContents();
  if (closedByPopstate) {
    closedByPopstate = false;
  } else if (history.state?.detail) {
    pageHistory.restore(history.state);
    pageHistory.push(pageHistory.state.url);
  }
  updateTitle(pageHistory.state.title, false);
});

const dialogCloseEl =
  dialog.querySelector("button") || dialog.querySelector("a.button");
dialogCloseEl.addEventListener("click", function (ev) {
  ev.preventDefault();
  dialog.close();
});

/**
 *
 * @param {URL} url
 */
async function getDetail(url) {
  try {
    const res = await fetch(url, {
      headers: {
        fetch: "true",
      },
    });
    updateTitle(res.headers.get("title"));

    // render errors sent as HTML as well as OK responses
    if (res.headers.get("content-type").startsWith("text/html")) {
      renderDetails(await res.text());
    } else {
      throw new Error(`${res.status} ${res.statusText}: ${await res.text()}`);
    }
  } catch (error) {
    console.error("fetch failed", error);
    renderDetails(new Text(error.message));
  }
}

/**
 * @param {MouseEvent} ev
 */
function handleDialogOpen(ev) {
  if (!(ev.ctrlKey || ev.metaKey || ev.shiftKey || ev.altKey)) {
    pageHistory.pushDetail(ev.target.href);
    getDetail(new URL(ev.target.href));
    ev.preventDefault();
  }
}

if (dialog.open) {
  dialog.removeAttribute("open");
  dialog.showModal();
}

addEventListener("popstate", function (ev) {
  if (dialog.open) {
    closedByPopstate = true;
    dialog.close();
  }
  if (ev.state?.detail) {
    const detailUrl = new URL(location.href);
    detailUrl.searchParams.delete("query");
    getDetail(detailUrl);
  } else if (ev.state != null) {
    pageHistory.restore(ev.state);
    document.title = pageHistory.state.title || siteTitle;
    if (pageHistory.state.results !== null) {
      queryInput.value = pageHistory.state.input;
      renderResults(pageHistory.state.results);
    }
  } else {
    document.title = siteTitle;
    resultsRange.deleteContents();
    search.reset();
  }
});
