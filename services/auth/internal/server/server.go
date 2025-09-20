package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/raulsilva-tech/e-commerce/services/auth/dto"
	"github.com/raulsilva-tech/e-commerce/services/auth/internal/db"
	"github.com/raulsilva-tech/e-commerce/services/auth/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Port            string
	DatabaseDSN     string
	RedisAddr       string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type Server struct {
	cfg            Config
	userRepository db.UserRepositoryInterface
	rdb            *redis.Client
}

func NewServer(cfg Config, userRepository *db.UserRepository) (http.Handler, error) {

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("warning: redis ping: %v", err)
	}

	s := &Server{
		cfg:            cfg,
		userRepository: userRepository,
		rdb:            rdb,
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

	user, err := s.userRepository.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
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

	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "name, email, password required", http.StatusBadRequest)
	}

	existingUser, err := s.userRepository.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {

		http.Error(w, "email already used", http.StatusConflict)
		return
	}

	user, err := entity.NewUser(0, req.Name, req.Email, req.Password, time.Now())
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	id, err := s.userRepository.Create(*user)
	if err != nil {
		fmt.Println("2", err.Error())
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	res := map[string]any{"id": id, "email": req.Email}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// helper for errors
func writeErr(w http.ResponseWriter, err error) {
	var st int
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		st = http.StatusGatewayTimeout
	default:
		st = http.StatusInternalServerError
	}
	http.Error(w, err.Error(), st)
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
		if username == "" || password == "" {
			http.Error(w, "username and password required", http.StatusBadRequest)
			return
		}
		user, err := s.userRepository.GetByEmail(username)
		if err != nil || user == nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		// generate tokens
		access, err := s.makeAccessToken(user.ID, user.Email)
		if err != nil {
			http.Error(w, "failed to create access token", http.StatusInternalServerError)
			return
		}
		refresh, err := s.makeRefreshToken()
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
		access, err := s.makeAccessToken(uid, "")
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

	ctx := context.Background()
	if err := s.rdb.Del(ctx, "refresh:"+payload.RefreshToken).Err(); err != nil {
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
func (s *Server) makeAccessToken(userID int64, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", userID),
		"iat":   now.Unix(),
		"exp":   now.Add(s.cfg.AccessTokenTTL).Unix(),
		"email": email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	return signed, err
}

func (s *Server) makeRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// -------------

func (s *Server) getConfig() Config {
	return s.cfg
}

func (s *Server) meHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("protected info"))
}
