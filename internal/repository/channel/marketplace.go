package channel

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/bpva/ad-marketplace/internal/entity"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func withFilters(b sq.SelectBuilder, filters []entity.Filter) sq.SelectBuilder {
	for _, f := range filters {
		b = b.Where(f)
	}
	return b
}

func (r *repo) GetChannels(
	ctx context.Context,
	filters []entity.Filter,
	sort entity.ChannelSort,
	limit, offset int,
) ([]entity.MVChannel, int, error) {
	countSQL, countArgs, err := withFilters(
		psql.Select("COUNT(*)").From("channel_marketplace"), filters,
	).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("building count query: %w", err)
	}

	countRows, err := r.db.Query(ctx, countSQL, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("counting channels: %w", err)
	}
	total, err := pgx.CollectOneRow(countRows, pgx.RowTo[int])
	if err != nil {
		return nil, 0, fmt.Errorf("scanning count: %w", err)
	}

	dataSQL, dataArgs, err := withFilters(
		psql.Select("*").From("channel_marketplace"), filters,
	).
		OrderBy(sort.OrderByClause()).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("building data query: %w", err)
	}

	rows, err := r.db.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying channels: %w", err)
	}

	channels, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[entity.MVChannel])
	if err != nil {
		return nil, 0, fmt.Errorf("scanning channels: %w", err)
	}

	return channels, total, nil
}

func (r *repo) RefreshMV(ctx context.Context) error {
	_, err := r.db.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY channel_marketplace")
	return err
}
