package cacher_test

import (
	"reflect"
	"testing"

	"pulley.com/shakesearch/pkg/cacher"
	"pulley.com/shakesearch/pkg/searcher"
)

func TestinMemory(t *testing.T) {
	cache := cacher.NewInMemoryCacher()

	key := "key"
	content := searcher.Response{
		Query: "Pully",
	}

	if _, err := cache.Get(key, searcher.Response{}); err == nil {
		t.Errorf("want no content")
	}

	cache.Set(key, content)

	res, err := cache.Get(key, searcher.Response{})
	if err != nil {
		t.Errorf("want no error")
	}

	if !reflect.DeepEqual(res, content) {
		t.Errorf("content not equal")
	}

}
