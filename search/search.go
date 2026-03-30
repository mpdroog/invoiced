// Package search provides full-text search across invoices, hours, and purchases.
package search

import (
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/idx"
	"github.com/mpdroog/invoiced/writer"
)

// SearchResponse contains search results.
type SearchResponse struct { //nolint:revive // maintaining public API
	Results []idx.SearchResult `json:"results"`
}

// Search handles search queries across invoices, hours, and purchases.
func Search(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	query := r.URL.Query().Get("q")

	if query == "" {
		if err := writer.Encode(w, r, SearchResponse{Results: []idx.SearchResult{}}); err != nil {
			log.Printf("search.Search encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	results, err := idx.Search(entity, query, 20)
	if err != nil {
		log.Printf("search.Search: %s", strconv.Quote(err.Error()))
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	if results == nil {
		results = []idx.SearchResult{}
	}

	if err := writer.Encode(w, r, SearchResponse{Results: results}); err != nil {
		log.Printf("search.Search encode: %s", strconv.Quote(err.Error()))
	}
}
