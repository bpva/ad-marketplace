package app

import (
	"net/http"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/bind"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

// HandleGetProfile returns user profile with settings
//
//	@Summary		Get user profile
//	@Tags			user
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.ProfileResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/user/profile [get]
func (a *App) HandleGetProfile() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/user/profile"))

	return func(w http.ResponseWriter, r *http.Request) {
		profile, err := a.user.GetProfile(r.Context())
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, profile)
	}
}

// HandleUpdateName updates user display name
//
//	@Summary		Update user name
//	@Tags			user
//	@Accept			json
//	@Security		BearerAuth
//	@Param			request	body	dto.UpdateNameRequest	true	"New name"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/user/name [patch]
func (a *App) HandleUpdateName() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/user/name"))

	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.UpdateNameRequest
		if err := bind.JSON(r, &req); err != nil {
			respond.Err(w, log, err)
			return
		}

		if err := a.user.UpdateName(r.Context(), req.Name); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

// HandleLinkWallet stores the user's TON wallet address
//
//	@Summary		Link TON wallet
//	@Tags			user
//	@Accept			json
//	@Security		BearerAuth
//	@Param			request	body	dto.LinkWalletRequest	true	"Wallet address"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/user/wallet [put]
func (a *App) HandleLinkWallet() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/user/wallet"))

	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.LinkWalletRequest
		if err := bind.JSON(r, &req); err != nil {
			respond.Err(w, log, err)
			return
		}

		if err := a.user.LinkWallet(r.Context(), req.Address); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

// HandleUnlinkWallet removes the user's TON wallet address
//
//	@Summary		Unlink TON wallet
//	@Tags			user
//	@Security		BearerAuth
//	@Success		204
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/user/wallet [delete]
func (a *App) HandleUnlinkWallet() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/user/wallet"))

	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.user.UnlinkWallet(r.Context()); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

// HandleUpdateSettings updates user settings (partial update)
//
//	@Summary		Update user settings
//	@Tags			user
//	@Accept			json
//	@Security		BearerAuth
//	@Param			request	body	dto.UpdateSettingsRequest	true	"Settings to update"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/user/settings [patch]
func (a *App) HandleUpdateSettings() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/user/settings"))

	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.UpdateSettingsRequest
		if err := bind.JSON(r, &req); err != nil {
			respond.Err(w, log, err)
			return
		}

		if err := a.user.UpdateSettings(r.Context(), req); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}
