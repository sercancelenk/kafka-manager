package service

type PaginatedResponse[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}
