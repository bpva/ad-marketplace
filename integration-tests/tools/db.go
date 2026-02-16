package tools

import "context"

func (t *Tools) Truncate(ctx context.Context, tables ...string) error {
	for _, table := range tables {
		if _, err := t.pool.Exec(ctx, "TRUNCATE TABLE "+table+" CASCADE"); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tools) TruncateAll(ctx context.Context) error {
	return t.Truncate(ctx, "deals", "posts", "channel_roles", "channels", "users")
}
