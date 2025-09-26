package webserver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/raulsilva-tech/e-commerce/services/auth/config"
	"github.com/raulsilva-tech/e-commerce/services/auth/dto"
	"github.com/raulsilva-tech/e-commerce/services/auth/internal/entity"
	"github.com/raulsilva-tech/e-commerce/services/auth/internal/usecase"
)

type Server struct {
	cfg         config.Config
	authUseCase usecase.AuthUseCase
	rdb         *redis.Client
}

func NewServer(cfg config.Config, uc usecase.AuthUseCase) (http.Handler, error) {

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("warning: redis ping: %v", err)
	}

	s := &Server{
		cfg:         cfg,
		authUseCase: uc,
		rdb:         rdb,
	}
	r := mux.NewRouter()
	r.HandleFunc("/health", s.healthHandler).Methods("GET")
	r.HandleFunc("/login", s.loginHandler).Methods("POST")
	r.HandleFunc("/signup", s.signupHandler).Methods("POST")
	r.HandleFunc("/oauth/token", s.tokenHandler).Methods("POST")
	r.HandleFunc("/logout", s.logoutHandler).Methods("POST")

	fs := http.FileServer(http.Dir("./docs"))
	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", fs))

	// example protected route using JWT middleware
	r.Handle("/me", s.jwtMiddleware(http.HandlerFunc(s.meHandler))).Methods("GET")
	return r, nil
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	_, err := s.authUseCase.Login(usecase.LoginInput{Email: req.Email, Password: req.Password})

	if err != nil {

		switch err {
		case entity.ErrEmailPasswordRequired:
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		case entity.ErrInvalidCredentials:
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	res := map[string]interface{}{"message": "use /oauth/token with grant_type=password to receive access and refresh tokens"}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) signupHandler(w http.ResponseWriter, r *http.Request) {

	var req dto.SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	id, err := s.authUseCase.Signup(usecase.SignupInput{Name: req.Name, Email: req.Email, Password: req.Password})
	if err != nil {

		switch err {
		case entity.ErrEmailPasswordRequired:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case entity.ErrEmailAlreadyUsed:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		return
	}

	res := map[string]any{"id": id, "email": req.Email}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// ---------------- OAuth2-like token endpoint ----------------

// token request for grant_type=password
// form values (application/x-www-form-urlencoded): grant_type=password&username=...&password=...
// for refresh: grant_type=refresh_token&refresh_token=...

func (s *Server) tokenHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	grant := r.FormValue("grant_type")
	switch grant {
	case "password":
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := s.authUseCase.Login(usecase.LoginInput{Email: username, Password: password})

		if err != nil {

			switch err {
			case entity.ErrEmailPasswordRequired:
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			case entity.ErrInvalidCredentials:
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		// generate tokens
		access, err := MakeAccessToken(user.ID, user.Email, s.cfg.AccessTokenTTL, s.cfg.JWTSecret)
		if err != nil {
			http.Error(w, "failed to create access token", http.StatusInternalServerError)
			return
		}
		refresh, err := MakeRefreshToken()
		if err != nil {
			http.Error(w, "failed to create refresh token", http.StatusInternalServerError)
			return
		}
		// persist refresh in redis
		ctx := context.Background()
		if err := s.rdb.Set(ctx, "refresh:"+refresh, fmt.Sprintf("%d", user.ID), s.cfg.RefreshTokenTTL).Err(); err != nil {
			log.Printf("warning: failed set refresh token: %v", err)
		}
		res := map[string]interface{}{
			"access_token":  access,
			"token_type":    "bearer",
			"expires_in":    int(s.cfg.AccessTokenTTL.Seconds()),
			"refresh_token": refresh,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
		return

	case "refresh_token":
		rToken := r.FormValue("refresh_token")
		if rToken == "" {
			http.Error(w, "refresh_token required", http.StatusBadRequest)
			return
		}
		ctx := context.Background()
		uidStr, err := s.rdb.Get(ctx, "refresh:"+rToken).Result()
		if err != nil {
			http.Error(w, "invalid refresh token", http.StatusUnauthorized)
			return
		}
		// parse user id
		var uid int64
		_, err = fmt.Sscanf(uidStr, "%d", &uid)
		if err != nil {
			http.Error(w, "invalid token", http.StatusInternalServerError)
			return
		}
		// create new access token
		access, err := MakeAccessToken(uid, "", s.cfg.AccessTokenTTL, s.cfg.JWTSecret)
		if err != nil {
			http.Error(w, "failed to create access token", http.StatusInternalServerError)
			return
		}
		res := map[string]interface{}{
			"access_token": access,
			"token_type":   "bearer",
			"expires_in":   int(s.cfg.AccessTokenTTL.Seconds()),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
		return

	default:
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
		return
	}
}

// ---------------- Logout / Revoke ----------------
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if payload.RefreshToken == "" {
		http.Error(w, "refresh_token required", http.StatusBadRequest)
		return
	}

	if err := s.rdb.Del(context.Background(), "refresh:"+payload.RefreshToken).Err(); err != nil {
		log.Printf("warning: failed delete refresh token: %v", err)
	}
	w.WriteHeader(http.StatusNoContent)

}

// ---- JWT Middleware
func (s *Server) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		var tok string
		fmt.Sscanf(auth, "Bearer %s", &tok)
		if tok == "" {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}
		// parse
		parsed, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(s.cfg.JWTSecret), nil
		})
		if err != nil || !parsed.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		// inject claims into context if needed
		next.ServeHTTP(w, r)
	})
}

// ---------------- Helpers: JWT and refresh token ----------------
func MakeAccessToken(userID int64, email string, accessTokenTTL time.Duration, jwtSecret string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", userID),
		"iat":   now.Unix(),
		"exp":   now.Add(accessTokenTTL).Unix(),
		"email": email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtSecret))
	return signed, err
}

func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// -------------

func (s *Server) meHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("protected info"))
}
