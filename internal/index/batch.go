package index

import (
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/document"
)

type Batcher struct {
	batch *bleve.Batch
	write *WriteIndex
	err   error
	n     int
	max   int
}

func (index *WriteIndex) NewBatcher() *Batcher {
	return &Batcher{
		batch: index.index.NewBatch(),
		write: index,
		max:   index.batchSize,
	}
}

func (b *Batcher) Size() int {
	return b.max
}

func (b *Batcher) Flush() error {
	b.flush()

	return b.err
}

func (b *Batcher) flush() {
	if b.n == 0 {
		return
	}

	b.write.log.Debug("flushing batch", "size", b.n)

	err := b.write.index.Batch(b.batch)
	if err != nil {
		b.err = &BatchError{fault.Wrap(err, fmsg.Withf("could not flush batch"))}

		return
	}

	b.Reset()
}

func (b *Batcher) Reset() {
	b.err = nil
	b.n = 0
	b.batch.Reset()
}

func (b *Batcher) countAndFlush() {
	if b.n++; b.n >= b.max {
		b.flush()
	}
}

func (b *Batcher) IndexAdvanced(doc *document.Document) error {
	err := b.batch.IndexAdvanced(doc)
	if err != nil {
		return fault.Wrap(err, fmsg.Withf("could not index document"))
	}

	b.countAndFlush()

	if b.err != nil {
		return b.err
	}

	return nil
}

func (b *Batcher) Delete(id string) {
	b.batch.Delete(id)
}
