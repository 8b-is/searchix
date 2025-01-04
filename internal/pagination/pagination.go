package pagination

type Pagination struct {
	total uint64

	From int

	Size,
	Current,
	Prev,
	Next int

	Needed bool
}

func New(page int, pageSize int) *Pagination {
	return &Pagination{
		Current: page,
		From:    (page - 1) * pageSize,
		Size:    pageSize,
	}
}

func (p *Pagination) SetResults(total uint64) {
	p.total = total
	p.Needed = p.total > uint64(p.Size)
	if uint64(p.Current*p.Size) <= p.total {
		p.Next = p.Current + 1
		p.Prev = p.Current - 1
	}
}
