package esp

import (
	"context"

	"github.com/xh-polaris/paginator-go"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type (
	EsPaginator struct {
		store paginator.Store
		opts  *paginator.PaginationOptions
	}
)

func NewEsPaginator(store paginator.Store, opts *paginator.PaginationOptions) *EsPaginator {
	opts.EnsureSafe()
	return &EsPaginator{
		store: store,
		opts:  opts,
	}
}

// MakeSortOptions 生成ID分页查询选项
func (p *EsPaginator) MakeSortOptions(ctx context.Context) ([]types.SortCombinations, []types.FieldValue, error) {
	if p.opts.LastToken != nil {
		err := p.store.LoadSorter(ctx, *p.opts.LastToken, *p.opts.Backward)
		if err != nil {
			return nil, nil, err
		}
	}

	sorter := p.store.GetSorter()
	sort, sa, err := sorter.(EsSorter).MakeSortOptions(*p.opts.Backward)
	if err != nil {
		return nil, nil, err
	}
	return sort, sa, nil
}

func (p *EsPaginator) StoreSorter(ctx context.Context, first, last any) error {
	token, err := p.store.StoreSorter(ctx, p.opts.LastToken, first, last)
	p.opts.LastToken = token
	return err
}
