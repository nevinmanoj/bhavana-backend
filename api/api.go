package api

type GetAllResponsePage[T any] struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	// TotalRecords, Limit and Offset are 0 for endpoints that don't paginate
	// (the caller never sent limit/offset, so Data is the full result set).
	TotalRecords int64 `json:"total_records"`
	Limit        int64 `json:"limit"`
	Offset       int64 `json:"offset"`
	Data         []T   `json:"data"`
}
type GetResponsePage[T any] struct {
	Message    string `json:"message"`
	Data       T      `json:"data"`
	StatusCode int    `json:"status_code"`
}

type PostResponsePage[T any] struct {
	Message    string `json:"message"`
	Data       T      `json:"data"`
	StatusCode int    `json:"status_code"`
}
type PutResponsePage[T any] struct {
	Message    string `json:"message"`
	Data       T      `json:"data"`
	StatusCode int    `json:"status_code"`
}
type DeleteResponsePage struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}
type ErrorResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}
