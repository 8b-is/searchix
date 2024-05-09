const search = document.getElementById("search");
let results = document.getElementById("results");
let pagination = document.getElementById("pagination");

const range = new Range();
range.setStartAfter(search);
range.setEndAfter(search.parentNode.lastChild);

let state = history.state || {
  url: new URL(location).toString(),
  fragment: range.cloneContents().innerHTML || "",
  opened: [],
};

let escapePolicy = null;
if (window.trustedTypes && trustedTypes.createPolicy) {
  escapePolicy = trustedTypes.createPolicy("fetch", {
    createHTML: (string) => string,
  });
}

function detailsToggled(ev) {
  const nextURL = new URL(location);
  if (ev.newState == "open" || ev.target.open === true) {
    state.opened.push(this.id);
    nextURL.hash = this.id;
  } else {
    state.opened = state.opened.filter((x) => x != this.id);
    nextURL.hash = "";
  }
  state.url = nextURL.toJSON();
  history.replaceState(state, "", nextURL);
}
function addToggleEventListeners(results) {
  results.querySelectorAll("details").forEach((details) =>
    // toggle event doesn't bubble :(
    details.addEventListener("toggle", detailsToggled, { passive: true }),
  );
}

function paginationLinkClicked(ev) {
  const url = new URL(ev.target.href);
  getResults(url);
  ev.preventDefault();
}

function addPaginationEventListeners(pagination) {
  pagination.addEventListener("click", paginationLinkClicked);
}

function renderFragmentHTML(html) {
  const fragment = range.createContextualFragment(
    escapePolicy !== null ? escapePolicy.createHTML(html) : html,
  );
  results = fragment.querySelector("#results");
  pagination = fragment.querySelector("#pagination");
  range.deleteContents();
  range.insertNode(fragment);
  addToggleEventListeners(results);
  addPaginationEventListeners(pagination);
}

function getResults(url) {
  fetch(url, {
    headers: {
      fetch: "true",
    },
  })
    .then(async function (res) {
      if (res.ok) {
        return res.text();
      } else {
        throw new Error(`${res.status} ${res.statusText}: ${await res.text()}`);
      }
    })
    .then(function (html) {
      state.fragment = html;
      history.pushState(state, null, url);
      return renderFragmentHTML(html);
    })
    .catch(function (error) {
      range.deleteContents();
      range.insertNode(new Text(error.message));
      console.error("fetch failed", error);
    })
    .finally(function () {
      state.url = url.toJSON();
      state.opened = [];
      history.pushState(state, null, url);
    });
}

search.addEventListener("submit", function (ev) {
  const url = new URL(this.action);
  url.search = new URLSearchParams(new FormData(this)).toString();
  getResults(url);
  ev.preventDefault();
});

if (results !== null) {
  addToggleEventListeners(results);
}
if (pagination !== null) {
  addPaginationEventListeners(pagination);
}

if (state.opened.length > 0) {
  state.opened.forEach((id) =>
    document.getElementById(id).setAttribute("open", "open"),
  );
} else if (location.hash) {
  document.getElementById(location.hash.slice(1)).setAttribute("open", "open");
}

addEventListener("popstate", function (ev) {
  if (ev.state != null) {
    url = new URL(ev.state.url);
    if (!url.pathname.endsWith("/search") && ev.state.fragment !== null) {
      renderFragmentHTML(ev.state.fragment);
      return;
    }
  }
  range.deleteContents();
  search.reset();
});
