package seeds

import (
	"context"
	"fmt"

	"github.com/bpva/ad-marketplace/internal/entity"
	settings_repo "github.com/bpva/ad-marketplace/internal/repository/settings"
	user_repo "github.com/bpva/ad-marketplace/internal/repository/user"
)

type seedUser struct {
	entity *entity.User
	mode   entity.PreferredMode
}

func (s *Seeder) seedUsers(ctx context.Context) ([]seedUser, error) {
	users := user_repo.New(s.db)
	settings := settings_repo.New(s.db)

	defs := []struct {
		tgID int64
		name string
		mode entity.PreferredMode
	}{
		{75775519, "Pavel", entity.PreferredModePublisher},
		{88888888, "Channel Owner", entity.PreferredModePublisher},
		{99999999, "Advertiser", entity.PreferredModeAdvertiser},
	}

	result := make([]seedUser, 0, len(defs))
	for _, d := range defs {
		u, err := users.Create(ctx, d.tgID, d.name)
		if err != nil {
			return nil, fmt.Errorf("create user %s: %w", d.name, err)
		}

		st, err := settings.Create(ctx, u.ID)
		if err != nil {
			return nil, fmt.Errorf("create settings for %s: %w", d.name, err)
		}
		st.PreferredMode = d.mode
		st.OnboardingFinished = true
		if err := settings.Update(ctx, st); err != nil {
			return nil, fmt.Errorf("update settings for %s: %w", d.name, err)
		}

		result = append(result, seedUser{entity: u, mode: d.mode})
	}
	return result, nil
}
