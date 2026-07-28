# Database Schema

PostgreSQL via GORM. All UUIDs default to `gen_random_uuid()`.

---

## `users`

```sql
id          UUID        PK, default gen_random_uuid()
email       TEXT        nullable
name        TEXT        nullable
avatar_url  TEXT        nullable
is_active   BOOLEAN     default true
is_deleted  BOOLEAN     default false
created_at  TIMESTAMP   autoCreateTime
```

---

## `profiles`

One-to-one with `users`. Stores optional anonymous identity info.

```sql
user_id    UUID        PK, FK → users.id (CASCADE)
nickname   TEXT        nullable
is_male    BOOLEAN     nullable
age        INTEGER     nullable
is_hidden  BOOLEAN     default false
updated_at TIMESTAMP   autoUpdateTime
```

---

## `rooms`

A chat session between exactly two users.

```sql
id         UUID        PK, default gen_random_uuid()
user1_id   UUID        FK → users.id (CASCADE)
user2_id   UUID        FK → users.id (CASCADE)
created_at TIMESTAMP   autoCreateTime
ended_at   TIMESTAMP   nullable — set when either user leaves
```

A room is **active** when `ended_at IS NULL`.

---

## `active_room_members`

Database-enforced ownership for active rooms. These rows are inserted in the
same transaction as the room and removed in the transaction that ends it.

```sql
user_id    UUID        PK, FK → users.id (CASCADE)
room_id    UUID        FK → rooms.id (CASCADE)
created_at TIMESTAMPTZ not null
```

The `rooms_active_membership_trigger` inserts/removes these rows whenever a room
is created, ended, or changes participants. Because `user_id` is the primary
key, PostgreSQL rejects any concurrent room creation that would place one user
in two active rooms, regardless of whether that user appears as `user1_id` or
`user2_id`. This also protects direct SQL writes that bypass the Go repository.

---

## `messages`

```sql
id         UUID        PK, default gen_random_uuid()
room_id    UUID        FK → rooms.id (CASCADE)
sender_id  UUID        FK → users.id (CASCADE)
content    TEXT        not null
created_at TIMESTAMP   autoCreateTime
```

Messages are deleted by a cleanup goroutine shortly after the room ends.

---

## `banned_words`

Words used by the moderation service to auto-report messages at send time.

```sql
id          UUID        PK
word        TEXT        unique, not null
category    TEXT        default 'General'
created_by  UUID        FK → users.id
created_at  TIMESTAMP   autoCreateTime
```

---

## `reports`

Auto-generated or user-submitted reports.

```sql
id               UUID        PK
reporter_id      UUID        FK → users.id
reported_user_id UUID        FK → users.id (CASCADE)
room_id          UUID        FK → rooms.id
status           TEXT        default 'pending'
created_at       TIMESTAMP   autoCreateTime
```

---

## Notes

- No `queues` table — the matchmaking queue is a Redis sorted set
  (`queue:waiting`, score = join timestamp ms). In-flight matches use
  token-owned, 30-second Redis reservations until room creation is finalized.
- `rooms` has no `category`, `is_sensitive`, or `is_deleted` columns.
- `profiles` has no `city` column — removed in an earlier refactor.

## Migrations

Schema changes are versioned SQL files in `api/pkg/database/migrations` and are
recorded in `schema_migrations`. Each migration runs in a transaction under a
PostgreSQL advisory lock. Run them explicitly before the API:

```bash
cd api
go run ./cmd/migrate
```

The API server no longer runs `AutoMigrate` during startup.
