const search = document.getElementById("search");
const nav = document.querySelectorAll("body > header > nav")[0];
const queryInput = document.getElementById("query");
const dialog = document.getElementById("dialog");
const results = document.getElementById("results");
let pagination = document.getElementById("pagination");

const resultsRange = new Range();
resultsRange.setStart(results, 0);
resultsRange.setEnd(results, results.childNodes.length);

const detailsRange = new Range();
detailsRange.setStartAfter(dialog.firstElementChild);
detailsRange.setEndAfter(dialog.lastElementChild);

let urlLocation = new URL(location);
let state = history.state || {
  url: urlLocation.toString(),
  input: urlLocation.searchParams.get("query"),
  results: resultsRange.cloneContents().innerHTML || null,
};

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
function paginationLinkClicked(ev) {
  const url = new URL(ev.target.href);
  getResults(url);
  ev.preventDefault();
}

/**
 *
 * @param {HTMLElement} pagination
 */
function addPaginationEventListeners(pagination) {
  Array.from(pagination.children).forEach((child) =>
    child.addEventListener("click", paginationLinkClicked),
  );
}

function renderResults(html) {
  const fragment = resultsRange.createContextualFragment(
    escapePolicy !== null ? escapePolicy.createHTML(html) : html,
  );
  pagination = fragment.querySelector("#pagination");
  resultsRange.deleteContents();
  resultsRange.insertNode(fragment);
  addOpenDialogListeners(results);
  if (pagination !== null) {
    addPaginationEventListeners(pagination);
  }
}

/**
 *
 * @param {URL} url
 */
async function getResults(url) {
  try {
    state.url = url.toJSON();
    history.pushState(state, null, url);
    const res = await fetch(url, {
      headers: {
        fetch: "true",
      },
    });

    // render errors sent as HTML as well as OK responses
    if (res.headers.get("content-type").startsWith("text/html")) {
      state.results = await res.text();
      history.replaceState(state, null, url);
      renderResults(state.results);
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
  state.input = formData.get("query");
  getResults(url);
  ev.preventDefault();
});

if (results !== null) {
  addOpenDialogListeners(results);
}
if (pagination !== null) {
  addPaginationEventListeners(pagination);
}

document.querySelector("a.current").addEventListener("click", function (ev) {
  search.reset();
  state.input = null;
  resultsRange.deleteContents();
  state.results = "";
  history.pushState(state, null, ev.target.href);
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

dialog.addEventListener("close", function (event) {
  detailsRange.deleteContents();
});

dialog.querySelector("button").addEventListener("click", function () {
  dialog.close();
});

/**
 *
 * @param {URL} url
 */
async function getDetail(url) {
  try {
    state.url = url.toJSON();
    const res = await fetch(url, {
      headers: {
        fetch: "true",
      },
    });

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
    getDetail(new URL(ev.target.href));
    ev.preventDefault();
  }
}

addEventListener("popstate", function (ev) {
  if (ev.state != null) {
    url = new URL(ev.state.url);
    if (ev.state.results !== null) {
      queryInput.value = ev.state.input;
      renderResults(ev.state.results);
    }
  } else {
    resultsRange.deleteContents();
    search.reset();
  }
});
