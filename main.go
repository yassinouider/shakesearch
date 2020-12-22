package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"pulley.com/shakesearch/pkg/cacher"
	"pulley.com/shakesearch/pkg/render"
	"pulley.com/shakesearch/pkg/searcher"
)

func main() {
	log.Println("new")
	dat, err := ioutil.ReadFile("./data/completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	suffixarraySearcher := searcher.NewSuffixArraySearcher(dat)
	inmemoryCacher := cacher.NewInMemoryCacher()
	handler := NewHandler(suffixarraySearcher, inmemoryCacher)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./public"))

	mux.Handle("/", fs)
	mux.HandleFunc("/search", handler.Search)

	fmt.Printf("Listening on port %s...\n", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}

type Handler struct {
	Searcher searcher.Searcher
	Render   render.Render
	Cacher   cacher.Cacher
}

func NewHandler(s searcher.Searcher, c cacher.Cacher) *Handler {
	return &Handler{
		Searcher: s,
		Render:   render.NewJsonRender(),
		Cacher:   c,
	}
}

func (h Handler) ChangerRender(r render.Render) {
	h.Render = r
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.Render.Error(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	if h.Cacher != nil {
		if cached, err := h.Cacher.Get(r.RequestURI); err == nil {
			log.Printf("content from cache for %s\n", r.RequestURI)
			h.Render.Success(w, cached)
			return
		}
	}

	searchRequest := SearchRequest{}
	if err := searchRequest.Bind(r); err != nil {
		h.Render.Error(w, http.StatusBadRequest, err)
		return
	}

	req := searcher.Request{
		Query:            searchRequest.Q,
		CaseSensitive:    searchRequest.Sensitive,
		ExactMatch:       searchRequest.ExactMatch,
		CharBeforeQuery:  searchRequest.Before,
		CharAfterQuery:   searchRequest.After,
		HighlightPreTag:  "<em>",
		HighlightPostTag: "</em>",
	}

	res, err := h.Searcher.Search(req)
	if err != nil {
		h.Render.Error(w, http.StatusBadRequest, err)
		return
	}

	if h.Cacher != nil {
		h.Cacher.Set(r.RequestURI, *res)
	}

	h.Render.Success(w, res)
}

type SearchRequest struct {
	Q          string
	Sensitive  bool
	ExactMatch bool
	Before     int
	After      int
}

func (sr *SearchRequest) Bind(r *http.Request) error {
	const (
		defaultSensitive = false
		defaultBefore    = 215
		defaultAfter     = 215
	)

	query := r.URL.Query()

	q := query.Get("q")
	if q == "" {
		return errors.New("missing search query in URL params")
	}

	if strings.HasPrefix(q, `"`) && strings.HasSuffix(q, `"`) {
		sr.ExactMatch = true
		q = strings.TrimPrefix(q, `"`)
		q = strings.TrimSuffix(q, `"`)
	}

	sr.Q = q

	sensitive, err := strconv.ParseBool(query.Get("sensitive"))
	if err != nil {
		sr.Sensitive = defaultSensitive
	} else {
		sr.Sensitive = sensitive
	}

	before, err := strconv.Atoi(query.Get("before"))
	if err != nil {
		sr.Before = defaultBefore
	} else {
		sr.Before = before
	}

	after, err := strconv.Atoi(query.Get("after"))
	if err != nil {
		sr.After = defaultAfter
	} else {
		sr.After = after
	}

	return nil

}
