package pagination

type Pagination struct {
	Total      int64  `json:"total"`
	PrevCursor string `json:"prevCursor,omitempty"`
	NextCursor string `json:"nextCursor,omitempty"`
	HasPrev    bool   `json:"hasPrev"`
	HasNext    bool   `json:"hasNext"`
}
