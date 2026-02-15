package seeds

import (
	"context"
	"fmt"

	post_repo "github.com/bpva/ad-marketplace/internal/repository/post"
)

func (s *Seeder) seedPosts(ctx context.Context, users []seedUser) error {
	posts := post_repo.New(s.db)

	defs := []struct {
		userIdx int
		text    string
	}{
		{0, "Check out our new product launch! Limited time offer."},
		{0, "Join our community for exclusive updates and deals."},
		{2, "Looking for channel partners to promote our app."},
	}

	for _, d := range defs {
		text := d.text
		_, err := posts.Create(
			ctx,
			users[d.userIdx].entity.ID,
			nil,
			&text,
			nil,
			nil,
			nil,
			false,
			false,
		)
		if err != nil {
			return fmt.Errorf("create post: %w", err)
		}
	}
	return nil
}
