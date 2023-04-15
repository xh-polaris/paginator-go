package mongop

import (
	"context"

	"github.com/xh-polaris/paginator-go"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	MongoPaginator struct {
		opts  *paginator.PaginationOptions
		store paginator.Store
	}
)

func NewMongoPaginator(store paginator.Store, opts *paginator.PaginationOptions) *MongoPaginator {
	opts.EnsureSafe()
	return &MongoPaginator{
		store: store,
		opts:  opts,
	}
}

// MakeSortOptions 生成ID分页查询选项，并将filter在原地更新
func (p *MongoPaginator) MakeSortOptions(ctx context.Context, filter bson.M) (bson.M, error) {
	if p.opts.LastToken != nil {
		err := p.store.LoadSorter(ctx, *p.opts.LastToken, *p.opts.Backward)
		if err != nil {
			return nil, err
		}
	}

	sorter := p.store.GetSorter()
	sort, err := sorter.(MongoSorter).MakeSortOptions(filter, *p.opts.Backward)
	if err != nil {
		return nil, err
	}
	return sort, nil
}

func (p *MongoPaginator) StoreSorter(ctx context.Context, first, last any) error {
	token, err := p.store.StoreSorter(ctx, p.opts.LastToken, first, last)
	p.opts.LastToken = token
	return err
}
