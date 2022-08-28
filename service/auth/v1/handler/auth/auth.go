package auth

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/util/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"wz2100.net/microlobby/service/auth/v1/argon2"
	"wz2100.net/microlobby/service/auth/v1/config"
	"wz2100.net/microlobby/service/auth/v1/db"
	proxyConfig "wz2100.net/microlobby/service/http_proxy/config"
	scomponent "wz2100.net/microlobby/service/settings/v1/component"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/proto/authservicepb/v1"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
	"wz2100.net/microlobby/shared/proto/userpb/v1"
	"wz2100.net/microlobby/shared/utils"
)

const pkgPath = config.PkgPath + "/handler/auth"

type Config struct {
	RefreshTokenExpiry int64 `json:"refresh_token_expiry"`
	AccessTokenExpiry  int64 `json:"access_token_expiry"`
}

type Handler struct {
	cRegistry *component.Registry
	settings  map[string][]byte
	config    *Config

	svcName string

	logrus component.LogrusComponent
}

func NewHandler(cregistry *component.Registry) (*Handler, error) {
	h := &Handler{
		cRegistry: cregistry,
		settings:  make(map[string][]byte),
		config:    &Config{},
		svcName:   cregistry.Service.Name(),
	}

	return h, nil
}

func (h *Handler) Start() error {
	logrus, err := component.Logrus(h.cRegistry)
	if err != nil {
		return errors.FromError(err)
	}
	h.logrus = logrus

	ctx := component.RegistryToContext(utils.CtxForService(context.Background()), h.cRegistry)
	s, err := scomponent.SettingsV1(h.cRegistry)
	if err != nil {
		return errors.FromError(err)
	}

	se, err := s.Get(ctx, "", "", h.svcName, "config")
	if err == nil {
		err = json.Unmarshal(se.Content, h.config)
	} else {
		h.config.RefreshTokenExpiry = 86400 * 14 // 14 days
		h.config.AccessTokenExpiry = 900         // 15 minutes
		craw, err := json.Marshal(h.config)
		if err != nil {
			return errors.FromError(err)
		}

		_, err = s.Upsert(ctx, &settingsservicepb.UpsertRequest{
			Service:     h.svcName,
			Name:        "config",
			Content:     craw,
			RolesRead:   []string{auth.ROLE_ADMIN, auth.ROLE_SERVICE},
			RolesUpdate: []string{auth.ROLE_ADMIN, auth.ROLE_SERVICE},
		})

		if err != nil {
			return errors.FromError(err)
		}
	}

	_, ok := h.settings[proxyConfig.SettingNameJWTRefreshTokenPub]
	_, ok2 := h.settings[proxyConfig.SettingNameJWTRefreshTokenPriv]
	if !ok || !ok2 {
		spub, epub := s.Get(ctx, "", "", proxyConfig.Name, proxyConfig.SettingNameJWTRefreshTokenPub)
		spri, epri := s.Get(ctx, "", "", proxyConfig.Name, proxyConfig.SettingNameJWTRefreshTokenPriv)

		if epub != nil || epri != nil {
			// Wait for http_proxy to generate the key
			time.Sleep(5 * time.Second)

			spub, epub = s.Get(ctx, "", "", proxyConfig.Name, proxyConfig.SettingNameJWTRefreshTokenPub)
			spri, epri = s.Get(ctx, "", "", proxyConfig.Name, proxyConfig.SettingNameJWTRefreshTokenPriv)
			if epub != nil || epri != nil {
				return errors.New(h.svcName, "Failed to get keys", http.StatusInternalServerError)
			}
		}

		h.settings[proxyConfig.SettingNameJWTRefreshTokenPub] = spub.Content
		h.settings[proxyConfig.SettingNameJWTRefreshTokenPriv] = spri.Content
	}

	_, ok = h.settings[proxyConfig.SettingNameJWTAccessTokenPub]
	_, ok2 = h.settings[proxyConfig.SettingNameJWTAccessTokenPriv]
	if !ok || !ok2 {
		spub, epub := s.Get(ctx, "", "", proxyConfig.Name, proxyConfig.SettingNameJWTAccessTokenPub)
		spri, epri := s.Get(ctx, "", "", proxyConfig.Name, proxyConfig.SettingNameJWTAccessTokenPriv)

		if epub != nil || epri != nil {
			// Wait for http_proxy to generate the key
			time.Sleep(5 * time.Second)

			spub, epub = s.Get(ctx, "", "", proxyConfig.Name, proxyConfig.SettingNameJWTAccessTokenPub)
			spri, epri = s.Get(ctx, "", "", proxyConfig.Name, proxyConfig.SettingNameJWTAccessTokenPriv)
			if epub != nil || epri != nil {
				return errors.New(h.svcName, "Failed to get keys", http.StatusInternalServerError)
			}
		}

		h.settings[proxyConfig.SettingNameJWTAccessTokenPub] = spub.Content
		h.settings[proxyConfig.SettingNameJWTAccessTokenPriv] = spri.Content
	}

	return nil
}

func (h *Handler) Stop() error {
	return nil
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

func (s *Handler) UserDelete(ctx context.Context, in *authservicepb.UserIDRequest, out *emptypb.Empty) error {
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
	if in.Username == auth.ROLE_SERVICE {
		return errors.New(s.svcName, "User already exists", http.StatusConflict)
	}

	hash, err := argon2.Hash(in.Password, argon2.DefaultParams)
	if err != nil {
		return err
	}

	result, err := db.UserCreate(ctx, in.Username, hash, in.Email, []string{auth.ROLE_USER})
	if err != nil {
		return errors.New(s.svcName, "User already exists", http.StatusConflict)
	}

	out.Id = result.ID.String()
	out.Email = result.Email
	out.Username = result.Username
	out.Roles = result.Roles

	return nil
}

func (s *Handler) genTokens(ctx context.Context, user *db.User, out *authservicepb.Token) error {
	pRefreshKey, ok := s.settings[proxyConfig.SettingNameJWTRefreshTokenPriv]
	if !ok {
		return errors.New(s.svcName, "no signinkey, can't generate the access token, check again later", http.StatusTooEarly)
	}
	pRefreshEDKey := ed25519.PrivateKey(pRefreshKey)

	// Create the Claims
	refreshClaims := &auth.JWTMicrolobbyClaims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + s.config.RefreshTokenExpiry,
			IssuedAt:  time.Now().Unix(),
			Issuer:    s.svcName,
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

	pAccessKey, ok := s.settings[proxyConfig.SettingNameJWTAccessTokenPriv]
	if !ok {
		return errors.New(s.svcName, "can't generate the access token, check again later", http.StatusTooEarly)
	}
	pAccessEDKey := ed25519.PrivateKey(pAccessKey)

	// Create the Claims
	accessClaims := &auth.JWTMicrolobbyClaims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + s.config.AccessTokenExpiry,
			IssuedAt:  time.Now().Unix(),
			Issuer:    s.svcName,
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

func (s *Handler) Login(ctx context.Context, in *authservicepb.LoginRequest, out *authservicepb.Token) error {
	user, err := db.UserFindByUsername(ctx, in.Username)
	if err != nil {
		log.Error(err)
		return errors.New(s.svcName, "Wrong username or password", http.StatusUnauthorized)
	}

	ok, err := argon2.Verify(in.Password, user.Password)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New(s.svcName, "Wrong username or password", http.StatusUnauthorized)
	}

	return s.genTokens(ctx, user, out)
}

func (s *Handler) Refresh(ctx context.Context, in *authservicepb.Token, out *authservicepb.Token) error {
	pRefreshPubKey, ok := s.settings[proxyConfig.SettingNameJWTRefreshTokenPub]
	if !ok {
		return errors.New(s.svcName, "can't check the refresh token, check again later", http.StatusTooEarly)
	}

	claims := auth.JWTMicrolobbyClaims{}
	_, err := jwt.ParseWithClaims(in.RefreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return ed25519.PublicKey(pRefreshPubKey), nil
	})
	if err != nil {
		return errors.New(s.svcName, fmt.Sprintf("checking the RefreshToken: %s", err), http.StatusBadRequest)
	}

	// Check claims (expiration)
	if err = claims.Valid(); err != nil {
		return fmt.Errorf("claims invalid: %s", err)
	}

	user, err := db.UserFindById(ctx, claims.Id)
	if err != nil {
		return errors.New(s.svcName, fmt.Sprintf("error fetching the user: %s", err), http.StatusUnauthorized)
	}

	return s.genTokens(ctx, user, out)
}
