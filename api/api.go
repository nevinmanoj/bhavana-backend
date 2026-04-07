package api

type GetAllResponsePage[T any] struct {
	Message      string `json:"message"`
	StatusCode   int    `json:"status_code"`
	TotalRecords int    `json:"total_records"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
	Data         []T    `json:"data"`
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
