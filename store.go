package paginator

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	suffixFront   = ":front"
	suffixBack    = ":back"
	defaultExpire = time.Minute * 5
)

type Store interface {
	GetSorter() any
	LoadSorter(ctx context.Context, lastToken string, backward bool) error
	StoreSorter(ctx context.Context, lastToken *string, first, last any) (*string, error)
}

type CacheStore struct {
	sorter     any
	sorterType reflect.Type
	cache      cache.Cache
	prefix     string
}

func NewCacheStore(c cache.Cache, sorter any, prefix string) *CacheStore {
	t := reflect.TypeOf(sorter)
	for t.Kind() == reflect.Interface || t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return &CacheStore{
		sorter:     sorter,
		sorterType: t,
		cache:      c,
		prefix:     prefix,
	}
}
func (s *CacheStore) GetSorter() any {
	return s.sorter
}

func (s *CacheStore) LoadSorter(ctx context.Context, lastToken string, backward bool) error {
	var key string
	if backward {
		key = s.prefix + lastToken + suffixFront
	} else {
		key = s.prefix + lastToken + suffixBack
	}
	s.sorter = reflect.New(s.sorterType).Interface()
	err := s.cache.GetCtx(ctx, key, s.sorter)
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheStore) StoreSorter(ctx context.Context, lastToken *string, first, last any) (*string, error) {
	if lastToken == nil {
		lastToken = new(string)
		*lastToken = uuid.New().String()
	}
	front := reflect.New(s.sorterType).Interface()
	err := copier.CopyWithOption(front, first, copier.Option{Converters: []copier.TypeConverter{{
		SrcType: primitive.ObjectID{},
		DstType: copier.String,
		Fn: func(src interface{}) (interface{}, error) {
			return src.(primitive.ObjectID).Hex(), nil
		},
	}}})
	if err != nil {
		return nil, err
	}
	//TODO 假如第一次成功，第二次失败会发生什么
	err = s.cache.SetWithExpireCtx(ctx, s.prefix+*lastToken+suffixFront, front, defaultExpire)
	if err != nil {
		return nil, err
	}

	back := reflect.New(s.sorterType).Interface()
	err = copier.CopyWithOption(back, last, copier.Option{Converters: []copier.TypeConverter{{
		SrcType: primitive.ObjectID{},
		DstType: copier.String,
		Fn: func(src interface{}) (interface{}, error) {
			return src.(primitive.ObjectID).Hex(), nil
		},
	}}})
	if err != nil {
		return nil, err
	}
	err = s.cache.SetWithExpireCtx(ctx, s.prefix+*lastToken+suffixBack, back, defaultExpire)
	if err != nil {
		return nil, err
	}
	return lastToken, nil
}

type RawStore struct {
	sorter     any
	sorterType reflect.Type
}

func NewRawStore(sorter any) *RawStore {
	t := reflect.TypeOf(sorter)
	for t.Kind() == reflect.Interface || t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return &RawStore{
		sorter:     sorter,
		sorterType: t,
	}
}

func (s *RawStore) GetSorter() any {
	return s.sorter
}

func (s *RawStore) LoadSorter(_ context.Context, lastToken string, backward bool) error {
	sorters := reflect.New(reflect.ArrayOf(2, reflect.PointerTo(s.sorterType)))
	err := json.Unmarshal([]byte(lastToken), sorters.Interface())
	if backward {
		s.sorter = sorters.Elem().Index(0).Interface()
	} else {
		s.sorter = sorters.Elem().Index(1).Interface()
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *RawStore) StoreSorter(_ context.Context, lastToken *string, first, last any) (*string, error) {
	bytes, err := json.Marshal([2]any{first, last})
	if err != nil {
		return nil, err
	}
	if lastToken == nil {
		lastToken = new(string)
	}
	*lastToken = string(bytes)
	return lastToken, nil
}
