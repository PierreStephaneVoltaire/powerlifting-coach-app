package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/powerlifting-coach-app/auth-service/internal/config"
)

type Service struct {
	client   *gocloak.GoCloak
	config   *config.Config
	adminToken *gocloak.JWT
	tokenExpiry time.Time
}

type UserClaims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	UserType string   `json:"user_type"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	UserType string `json:"user_type" binding:"required,oneof=athlete coach"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	UserType string `json:"user_type"`
}

type AuthResponse struct {
	Tokens TokenResponse `json:"tokens"`
	User   UserInfo      `json:"user"`
}

func NewService(cfg *config.Config) *Service {
	client := gocloak.NewClient(cfg.KeycloakURL)
	return &Service{
		client: client,
		config: cfg,
	}
}

func (s *Service) getAdminToken(ctx context.Context) (*gocloak.JWT, error) {
	if s.adminToken != nil && time.Now().Before(s.tokenExpiry) {
		return s.adminToken, nil
	}

	token, err := s.client.LoginAdmin(ctx, s.config.KeycloakAdminUser, s.config.KeycloakAdminPassword, "master")
	if err != nil {
		return nil, fmt.Errorf("failed to get admin token: %w", err)
	}

	s.adminToken = token
	s.tokenExpiry = time.Now().Add(time.Duration(token.ExpiresIn-60) * time.Second)

	return token, nil
}

func (s *Service) setUserPassword(ctx context.Context, accessToken, realm, userID, password string, temporary bool) error {
	client := resty.New()
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/reset-password", s.config.KeycloakURL, realm, userID)

	payload := map[string]interface{}{
		"type":      "password",
		"value":     password,
		"temporary": temporary,
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Put(url)

	if err != nil {
		return fmt.Errorf("failed to call reset password API: %w", err)
	}

	if resp.StatusCode() != 204 {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	adminToken, err := s.getAdminToken(ctx)
	if err != nil {
		return nil, err
	}

	user := gocloak.User{
		Email:         &req.Email,
		Username:      &req.Email,
		FirstName:     &req.Name,
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(true),
		Attributes: &map[string][]string{
			"user_type": {req.UserType},
		},
	}

	userID, err := s.client.CreateUser(ctx, adminToken.AccessToken, s.config.KeycloakRealm, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	err = s.setUserPassword(ctx, adminToken.AccessToken, s.config.KeycloakRealm, userID, req.Password, false)
	if err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	roleNames := []string{req.UserType}
	if req.UserType == "athlete" {
		roleNames = append(roleNames, "user")
	} else if req.UserType == "coach" {
		roleNames = append(roleNames, "user", "coach")
	}

	for _, roleName := range roleNames {
		role, err := s.client.GetRealmRole(ctx, adminToken.AccessToken, s.config.KeycloakRealm, roleName)
		if err != nil {
			continue
		}

		err = s.client.AddRealmRoleToUser(ctx, adminToken.AccessToken, s.config.KeycloakRealm, userID, []gocloak.Role{*role})
		if err != nil {
			return nil, fmt.Errorf("failed to assign role %s: %w", roleName, err)
		}
	}

	authResp, err := s.Login(ctx, LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	// Use the user ID from registration, not from the token
	authResp.User.ID = userID

	return authResp, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	token, err := s.client.Login(ctx, s.config.KeycloakClientID, s.config.KeycloakSecret, s.config.KeycloakRealm, req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Get user info from token
	claims, err := s.ValidateToken(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &AuthResponse{
		Tokens: TokenResponse{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresIn:    token.ExpiresIn,
			TokenType:    "Bearer",
		},
		User: UserInfo{
			ID:       claims.UserID,
			Email:    claims.Email,
			Name:     claims.Name,
			UserType: claims.UserType,
		},
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	token, err := s.client.RefreshToken(ctx, refreshToken, s.config.KeycloakClientID, s.config.KeycloakSecret, s.config.KeycloakRealm)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	return &TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*UserClaims, error) {
	_, claims, err := s.client.DecodeAccessToken(ctx, tokenString, s.config.KeycloakRealm)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	userClaims := &UserClaims{}
	
	if sub, ok := (*claims)["sub"].(string); ok {
		userClaims.UserID = sub
	}
	
	if email, ok := (*claims)["email"].(string); ok {
		userClaims.Email = email
	}
	
	if name, ok := (*claims)["name"].(string); ok {
		userClaims.Name = name
	}

	if realmAccess, ok := (*claims)["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			userClaims.Roles = make([]string, len(roles))
			for i, role := range roles {
				if roleStr, ok := role.(string); ok {
					userClaims.Roles[i] = roleStr
				}
			}
		}
	}

	if customClaims, ok := (*claims)["user_type"].(string); ok {
		userClaims.UserType = customClaims
	} else {
		for _, role := range userClaims.Roles {
			if role == "athlete" || role == "coach" {
				userClaims.UserType = role
				break
			}
		}
	}

	if exp, ok := (*claims)["exp"].(float64); ok {
		userClaims.ExpiresAt = jwt.NewNumericDate(time.Unix(int64(exp), 0))
	}

	return userClaims, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.client.Logout(ctx, s.config.KeycloakClientID, s.config.KeycloakSecret, s.config.KeycloakRealm, refreshToken)
}

func (s *Service) GetUserInfo(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	userInfo, err := s.client.GetUserInfo(ctx, tokenString, s.config.KeycloakRealm)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	result := make(map[string]interface{})
	userInfoBytes, _ := json.Marshal(userInfo)
	json.Unmarshal(userInfoBytes, &result)
	
	return result, nil
}