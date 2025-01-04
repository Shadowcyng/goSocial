package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit     int      `json:"limit" validate:"gte=1,lte=20"`
	Offset    int      `json:"offset" validate:"gte=0"`
	SortOrder string   `json:"sort_order" validate:"oneof=asc desc"`
	SortBy    string   `json:"sort_by" validate:"oneof=title content created_at"`
	Tags      []string `json:"tags" validate:"max=5"`
	Search    string   `json:"search" validate:"max=100"`
	Since     string   `json:"since" validate:"max=100"`
	Until     string   `json:"until" validate:"max=100"`
}

func (fq PaginatedFeedQuery) Parse(r *http.Request) PaginatedFeedQuery {
	qs := r.URL.Query()
	limit := qs.Get("limit")
	tags := qs.Get("tags")
	offset := qs.Get("offset")
	sortBy := qs.Get("sort_by")
	sortOrder := qs.Get("sort_order")
	search := qs.Get("search")
	since := qs.Get("since")
	until := qs.Get("until")

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return fq
		}
		fq.Limit = l
	}
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return fq
		}
		fq.Offset = o
	}
	if sortOrder != "" {
		fq.SortOrder = sortOrder
	}
	if sortBy != "" {
		fq.SortBy = sortBy
	}
	if search != "" {
		fq.Search = search
	}
	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	}
	if since != "" {
		fq.Since = parseTime(since)
	}
	if until != "" {
		fq.Until = parseTime(until)
	}

	return fq
}

func parseTime(str string) string {
	t, err := time.Parse(time.DateTime, str)
	if err != nil {
		return ""
	}
	return t.Format(time.DateTime)
}
