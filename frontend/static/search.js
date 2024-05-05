const search = document.getElementById("search");
const results = document.getElementById("results");
let state = history.state || {
  url: new URL(location).toJSON(),
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
function addToggleEventListeners() {
  results.querySelectorAll("details").forEach((details) =>
    // toggle event doesn't bubble :(
    details.addEventListener("toggle", detailsToggled, { passive: true }),
  );
}
search.addEventListener("submit", function (ev) {
  const url = new URL(this.action);
  url.search = new URLSearchParams(new FormData(this)).toString();
  fetch(url, {
    headers: {
      fetch: "true",
    },
  })
    .then(function (res) {
      state.url = url.toJSON();
      state.opened = [];
      history.pushState(state, null, url);
      if (res.ok) {
        return res.text();
      } else {
        throw new Error(res.statusText);
      }
    })
    .then(function (html) {
      results.innerHTML =
        escapePolicy !== null ? escapePolicy.createHTML(html) : html;
      addToggleEventListeners();
    })
    .catch(function (error) {
      console.error("fetch failed", error);
    });
  ev.preventDefault();
});

addToggleEventListeners();

if (state.opened.length > 0) {
  state.opened.forEach((id) =>
    document.getElementById(id).setAttribute("open", "open"),
  );
} else if (location.hash) {
  document.getElementById(location.hash.slice(1)).setAttribute("open", "open");
}

addEventListener("popstate", function (ev) {
  if (ev.state == null || ev.state.url.pathname == "/") {
    results.replaceChildren();
  }
});
