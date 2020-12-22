package cacher

import "pulley.com/shakesearch/pkg/searcher"

type Cacher interface {
	Get(key string) (searcher.Response, error)
	Set(key string, content searcher.Response)
}
