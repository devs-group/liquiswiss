package models

type Pagination struct {
	CurrentPage    int64 `json:"currentPage"`
	TotalCount     int64 `json:"totalCount"`
	TotalPages     int64 `json:"totalPages"`
	TotalRemaining int64 `json:"totalRemaining"`
}
