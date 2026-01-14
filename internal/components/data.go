package components

import (
	"net/url"
	"slices"
	"strconv"

	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/config"
	search "alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/pagination"
)

type TemplateData struct {
	Sources       []config.Source
	Source        config.Source
	Query         string
	ExtraHeadHTML string
	Code          int
	Message       string
	Title         string
	Assets        *frontend.AssetCollection
	SearchNav     SearchNav
}

type ResultData struct {
	TemplateData
	Query        string
	FacetQueries url.Values
	Results      *search.Result
}

type SearchNav struct {
	URL url.URL

	Prev string
	Next string
	All  string

	pagination *pagination.Pagination
}

func NewSearchNav(u url.URL) SearchNav {
	return SearchNav{URL: u}
}

func (u SearchNav) WithPagination(page *pagination.Pagination) SearchNav {
	u.pagination = page

	q := u.URL.Query()
	urlBase := u.URL.Path

	if page.Next != 0 {
		q.Set("page", strconv.Itoa(page.Next))
		u.Next = urlBase + "?" + q.Encode()
	}

	if page.Prev != 0 {
		if page.Prev == 1 {
			q.Del("page")
		} else {
			q.Set("page", strconv.Itoa(page.Prev))
		}
		u.Prev = urlBase + "?" + q.Encode()
	}

	if page.CanShowAll {
		q.Set("page", "0")
		u.All = urlBase + "?" + q.Encode()
	}

	return u
}

func (u SearchNav) HasFacet(field string, term string) bool {
	return slices.Contains(u.URL.Query()[field], term)
}

func (u SearchNav) AddFacet(field string, term string) string {
	nu := u.URL
	q := nu.Query()
	q.Add(field, term)
	nu.RawQuery = q.Encode()

	return nu.String()
}

func (u SearchNav) RemoveFacet(field string, term string) string {
	nu := u.URL
	q := nu.Query()
	q[field] = slices.DeleteFunc(q[field], func(v string) bool {
		return v == term
	})
	nu.RawQuery = q.Encode()

	return nu.String()
}
