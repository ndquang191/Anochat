# Login & Auth Flow

## Token Design

| Cookie | Type | Path | Expiry | Purpose |
|--------|------|------|--------|---------|
| `access_token` | HTTP-only, Secure | `/` | short (configured) | JWT for API auth |
| `refresh_token` | HTTP-only, Secure | `/auth` | long (configured) | Rotates access token |
| `temp_user_data` | readable | `/` | 60 s | Carries user info to frontend callback page |

---

## OAuth Flow

```
Browser              Frontend           Backend              Google
  |                     |                  |                    |
  | click Login         |                  |                    |
  |────────────────────►|                  |                    |
  |                     | GET /auth/google |                    |
  |                     |─────────────────►|                    |
  |◄────────────────────────────────────── redirect (302)       |
  | follow redirect     |                  |                    |
  |──────────────────────────────────────────────────────────►  |
  |◄────────────────────────────────────────────────────────── oauth consent
  | POST code to callback                  |                    |
  |──────────────────────────────────────►|                    |
  |                     |                  | exchange code      |
  |                     |                  |────────────────────►
  |                     |                  |◄──────── tokens ───|
  |                     |                  | get userinfo       |
  |                     |                  | get/create user    |
  |                     |                  | set access_token cookie
  |                     |                  | set refresh_token cookie
  |                     |                  | set temp_user_data cookie (60s)
  |◄──────────────────────────────────── redirect /callback (302)
  | GET /callback       |                  |
  |────────────────────►|                  |
  |                     | read temp_user_data cookie
  |                     | redirect → /
```

---

## Token Refresh

```
Frontend                     Backend                      Redis
   |                             |                           |
   | POST /auth/refresh          |                           |
   | (refresh_token cookie)      |                           |
   |────────────────────────────►|                           |
   |                             | hash(token) → key         |
   |                             | GET refresh:{hash} ──────►|
   |                             |◄────────────── userID ─── |
   |                             | check user is active      |
   |                             | generate new JWT          |
   |                             | set access_token cookie   |
   |◄────────────────────────────|                           |
   | 200 OK                      |                           |
```

The refresh token itself is **not rotated** — the same refresh token stays valid until logout or expiry.

---

## Logout

```
POST /auth/logout
  → hash(refresh_token) → DEL refresh:{hash} in Redis
  → clear access_token cookie
  → clear refresh_token cookie
```

---

## Middleware

`AuthMiddleware` checks (in order):
1. `Authorization: Bearer <token>` header
2. `access_token` cookie

On failure: `401` with Vietnamese user-facing message.

If the user's `is_active = false` (banned): `401` account suspended.

---

## Birthday Auto-Fill

After OAuth callback, the backend makes a best-effort call to the Google People API to fetch the user's birth year and compute age. This is non-blocking — if it fails, the profile just has no age set.
