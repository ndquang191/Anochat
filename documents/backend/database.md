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

- No `queues` table — the matchmaking queue is a Redis sorted set (`queue:waiting`, score = join timestamp ms).
- `rooms` has no `category`, `is_sensitive`, or `is_deleted` columns.
- `profiles` has no `city` column — removed in an earlier refactor.
