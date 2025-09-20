# Auth Service (Go) - OAuth2-like Example


Service features:
- `POST /signup` - create user (email/password)
- `POST /oauth/token` - get tokens using `grant_type=password` or `grant_type=refresh_token`
- `POST /logout` - revoke refresh token
- Access tokens are JWT (stateless): resource servers only need the `JWT_SECRET` to validate
- Refresh tokens are stored in Redis (stateful) and can be revoked


## Run locally (prereqs)
- Postgres and Redis up (use project `infra/docker-compose.yml`)


Build and run:
```bash
cd services/auth
go build -o auth ./cmd/server
POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/appdb?sslmode=disable" REDIS_ADDR="localhost:6379" JWT_SECRET="very-secret" ./auth
```


## DB migration
Apply `migrations/001_create_users.sql` to your Postgres DB.


## Example flows (curl)
### Signup
```bash
curl -X POST http://localhost:8080/signup \
-H 'Content-Type: application/json' \
-d '{"email":"user@example.com","password":"secret"}'
```


### Get tokens (Resource Owner Password Credentials grant)
```bash
curl -X POST http://localhost:8080/oauth/token \
-H 'Content-Type: application/x-www-form-urlencoded' \
-d 'grant_type=password&username=user@example.com&password=secret'
```
Response includes `access_token`, `refresh_token`, `expires_in`.


### Use access token
```bash
curl -H "Authorization: Bearer <ACCESS_TOKEN>" http://localhost:8080/me
```


### Refresh access token
```bash
curl -X POST http://localhost:8080/oauth/token \
-H 'Content-Type: application/x-www-form-urlencoded' \
-d 'grant_type=refresh_token&refresh_token=<REFRESH_TOKEN>'
```


### Logout / revoke refresh
```bash
curl -X POST http://localhost:8080/logout \
-H "Content-Type: application/json" \
-d '{"refresh_token":"<REFRESH_TOKEN>"}'
```