package entity

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type ChannelSortBy string

const (
	ChannelSortBySubscribers ChannelSortBy = "subscribers"
	ChannelSortByViews       ChannelSortBy = "views"
)

type Filter struct {
	Name  string
	Value any
}

func (f Filter) ToSql() (string, []any, error) { //nolint:revive
	switch f.Name {
	case "fulltext":
		q, _ := f.Value.(string)
		pattern := "%" + q + "%"
		return "(title ILIKE ? OR COALESCE(username, '') ILIKE ?)", []any{pattern, pattern}, nil
	case "has_ad_formats":
		return "ad_formats IS NOT NULL", nil, nil
	default:
		return "", nil, fmt.Errorf("unknown filter: %s", f.Name)
	}
}

var _ sq.Sqlizer = Filter{}

type ChannelSort struct {
	By    ChannelSortBy
	Order SortOrder
}

func (s ChannelSort) OrderByClause() string {
	dir := "ASC"
	if s.Order == SortOrderDesc {
		dir = "DESC"
	}

	if s.By == ChannelSortByViews {
		return fmt.Sprintf("avg_daily_views_7d %s NULLS LAST", dir)
	}

	return fmt.Sprintf("subscribers %s NULLS LAST", dir)
}
