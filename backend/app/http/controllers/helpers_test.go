package controllers

import "testing"

func TestPaginationMeta(t *testing.T) {
	cases := []struct {
		name         string
		page         int
		perPage      int
		total        int64
		wantLastPage int
	}{
		{"exact multiple", 1, 10, 100, 10},
		{"partial last page rounds up", 1, 10, 105, 11},
		{"single partial page", 1, 15, 3, 1},
		{"no rows still has one page", 1, 15, 0, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			meta := paginationMeta(c.page, c.perPage, c.total)
			if got := meta["last_page"]; got != c.wantLastPage {
				t.Errorf("last_page = %v, want %v", got, c.wantLastPage)
			}
			if got := meta["total"]; got != c.total {
				t.Errorf("total = %v, want %v", got, c.total)
			}
			if got := meta["per_page"]; got != c.perPage {
				t.Errorf("per_page = %v, want %v", got, c.perPage)
			}
		})
	}
}
