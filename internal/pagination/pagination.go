package pagination

import "alin.ovh/searchix/internal/config"

type Pagination struct {
	total uint64

	From int

	Size,
	Current,
	Prev,
	Next int

	Needed     bool
	CanShowAll bool
}

func New(page int, pageSize int) *Pagination {
	return &Pagination{
		Current: page,
		From:    (page - 1) * pageSize,
		Size:    pageSize,
		Needed:  page > 1,
	}
}

func (p *Pagination) SetResults(total uint64) {
	p.total = total
	p.Needed = p.Needed || (total > uint64(p.Size))
	p.CanShowAll = p.Needed && total < config.MaxResultsShowAll
	if p.Current > 1 {
		p.Prev = p.Current - 1
	}
	if uint64(p.Current*p.Size) <= p.total {
		p.Next = p.Current + 1
	}
}
