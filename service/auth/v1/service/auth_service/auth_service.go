package authService

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/golang/protobuf/ptypes/empty"
	"wz2100.net/microlobby/service/auth/v1/db"
	"wz2100.net/microlobby/shared/argon2"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/defs"
	"wz2100.net/microlobby/shared/proto/authservicepb/v1"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
	"wz2100.net/microlobby/shared/proto/userpb/v1"
)

type Config struct {
	RefreshTokenExpiry int64 `json:"refresh_token_expiry"`
	AccessTokenExpiry  int64 `json:"access_token_expiry"`
}

type Handler struct {
	cRegistry *component.Registry
	settings  map[string][]byte
	config    *Config
}

func NewHandler(cregistry *component.Registry) (*Handler, error) {
	h := &Handler{
		cRegistry: cregistry,
		settings:  make(map[string][]byte),
	}

	go func() {
		ctx := context.Background()
		s, err := component.SettingsV1(cregistry)
		if err != nil {
			panic(err)
		}

		_, ok := h.settings["config"]
		if !ok {
			var c *Config
			se, err := s.Get(ctx, "", "", defs.ServiceAuthV1, "config")
			if err == nil {
				err = json.Unmarshal(se.Content, c)
			}

			if err != nil {
				c = &Config{
					RefreshTokenExpiry: 900,        // 15 minutes
					AccessTokenExpiry:  86400 * 14, // 14 days
				}
				craw, err := json.Marshal(c)
				if err != nil {
					panic(err)
				}

				_, err = s.Upsert(ctx, &settingsservicepb.CreateRequest{
					Service:     defs.ServiceAuthV1,
					Name:        "config",
					Content:     craw,
					RolesRead:   []string{auth.ROLE_ADMIN},
					RolesUpdate: []string{auth.ROLE_ADMIN},
				})

				if err != nil {
					panic(err)
				}
			}

			h.config = c
		}

		_, ok = h.settings[defs.SettingNameJWTRefreshTokenPub]
		_, ok2 := h.settings[defs.SettingNameJWTRefreshTokenPriv]
		if !ok || !ok2 {
			spub, epub := s.Get(ctx, "", "", defs.ServiceHttpProxy, defs.SettingNameJWTRefreshTokenPub)
			spri, epri := s.Get(ctx, "", "", defs.ServiceHttpProxy, defs.SettingNameJWTRefreshTokenPriv)

			if epub != nil || epri != nil {
				// Wait for http_proxy to generate the key
				time.Sleep(5 * time.Second)

				spub, epub = s.Get(ctx, "", "", defs.ServiceHttpProxy, defs.SettingNameJWTRefreshTokenPub)
				spri, epri = s.Get(ctx, "", "", defs.ServiceHttpProxy, defs.SettingNameJWTRefreshTokenPriv)
				if epub != nil || epri != nil {
					panic(epub)
				}
			}

			h.settings[defs.SettingNameJWTRefreshTokenPub] = spub.Content
			h.settings[defs.SettingNameJWTRefreshTokenPriv] = spri.Content
		}

		_, ok = h.settings[defs.SettingNameJWTAccessTokenPub]
		_, ok2 = h.settings[defs.SettingNameJWTAccessTokenPriv]
		if !ok || !ok2 {
			spub, epub := s.Get(ctx, "", "", defs.ServiceHttpProxy, defs.SettingNameJWTAccessTokenPub)
			spri, epri := s.Get(ctx, "", "", defs.ServiceHttpProxy, defs.SettingNameJWTAccessTokenPriv)

			if epub != nil || epri != nil {
				// Wait for http_proxy to generate the key
				time.Sleep(5 * time.Second)

				spub, epub = s.Get(ctx, "", "", defs.ServiceHttpProxy, defs.SettingNameJWTAccessTokenPub)
				spri, epri = s.Get(ctx, "", "", defs.ServiceHttpProxy, defs.SettingNameJWTAccessTokenPriv)
				if epub != nil || epri != nil {
					panic(epub)
				}
			}

			h.settings[defs.SettingNameJWTAccessTokenPub] = spub.Content
			h.settings[defs.SettingNameJWTAccessTokenPriv] = spri.Content
		}
	}()

	return h, nil
}

func (s *Handler) UserList(ctx context.Context, in *authservicepb.ListRequest, out *authservicepb.UserListResponse) error {
	results, err := db.UserList(ctx, in.Limit, in.Offset)
	if err != nil {
		return err
	}

	// Copy the data to the result
	for _, result := range results {
		out.Data = append(out.Data, &authservicepb.UserListResponse_User{
			Id:       result.ID.String(),
			Username: result.Username,
			Email:    result.Email,
		})
	}

	return nil
}

func (s *Handler) UserDetail(ctx context.Context, in *authservicepb.UserIDRequest, out *userpb.User) error {
	result, err := db.UserDetail(ctx, in.UserId)
	if err != nil {
		return err
	}

	out.Id = result.ID.String()
	out.Email = result.Email
	out.Username = result.Username
	out.Roles = result.Roles

	return nil
}

func (s *Handler) UserDelete(ctx context.Context, in *authservicepb.UserIDRequest, out *empty.Empty) error {
	err := db.UserDelete(ctx, in.UserId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Handler) UserUpdateRoles(ctx context.Context, in *authservicepb.UpdateRolesRequest, out *userpb.User) error {
	result, err := db.UserUpdateRoles(ctx, in.UserId, in.Roles)
	if err != nil {
		return err
	}

	out.Id = result.ID.String()
	out.Email = result.Email
	out.Username = result.Username
	out.Roles = result.Roles

	return nil
}

func (s *Handler) Register(ctx context.Context, in *authservicepb.RegisterRequest, out *userpb.User) error {
	hash, err := argon2.Hash(in.Password, argon2.DefaultParams)
	if err != nil {
		return err
	}

	result, err := db.UserCreate(ctx, in.Username, hash, []string{auth.ROLE_USER})
	if err != nil {
		return err
	}

	out.Id = result.ID.String()
	out.Email = result.Email
	out.Username = result.Username
	out.Roles = result.Roles

	return nil
}

func (s *Handler) Login(ctx context.Context, in *authservicepb.LoginRequest, out *authservicepb.Token) error {
	user, err := db.UserFindByUsername(ctx, in.Username)
	if err != nil {
		return err
	}

	ok, err := argon2.Verify(in.Password, user.Password)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("http 403 - wrong username or password")
	}

	pRefreshKey, ok := s.settings[defs.SettingNameJWTRefreshTokenPriv]
	if !ok {
		return errors.New("no sigingkey, can't generate the token, it's my fault")
	}
	pRefreshEDKey := ed25519.PrivateKey(pRefreshKey)

	// Create the Claims
	refreshClaims := &auth.JWTMicrolobbyClaims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + s.config.RefreshTokenExpiry,
			IssuedAt:  time.Now().Unix(),
			Issuer:    defs.ServiceAuthV1,
			Id:        user.ID.String(),
			Subject:   user.Username,
		},
	}
	if err := refreshClaims.Valid(); err != nil {
		return err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodEdDSA, refreshClaims)
	refreshSignedToken, err := refreshToken.SignedString(pRefreshEDKey)
	if err != nil {
		return err
	}

	pAccessKey, ok := s.settings[defs.SettingNameJWTAccessTokenPriv]
	if !ok {
		return errors.New("no sigingkey, can't generate the token, it's my fault")
	}
	pAccessEDKey := ed25519.PrivateKey(pAccessKey)

	// Create the Claims
	accessClaims := &auth.JWTMicrolobbyClaims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + s.config.RefreshTokenExpiry,
			IssuedAt:  time.Now().Unix(),
			Issuer:    defs.ServiceAuthV1,
			Id:        user.ID.String(),
			Subject:   user.Username,
		},
		Roles: user.Roles,
	}
	if err := accessClaims.Valid(); err != nil {
		return err
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodEdDSA, accessClaims)
	accessSignedToken, err := accessToken.SignedString(pAccessEDKey)
	if err != nil {
		return err
	}

	out.RefreshToken = refreshSignedToken
	out.RefreshTokenExpiresAt = refreshClaims.ExpiresAt
	out.AccessToken = accessSignedToken
	out.AccessTokenExpiresAt = accessClaims.ExpiresAt

	return nil
}
