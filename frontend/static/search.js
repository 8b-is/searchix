const search = document.getElementById("search");
const results = document.getElementById("results");
search.addEventListener("submit", function (ev) {
  const url = new URL(this.action);
  url.search = new URLSearchParams(new FormData(this)).toString();
  const res = fetch(url, {
    headers: {
      fetch: "true",
    },
  })
    .then(function (res) {
      window.history.pushState(null, null, url);
      if (res.ok) {
        return res.text();
      } else {
        throw new Error(res.statusText);
      }
    })
    .then(function (html) {
      results.innerHTML = html;
    })
    .catch(function (error) {
      console.error("fetch failed", error);
    });
  ev.preventDefault();
});
