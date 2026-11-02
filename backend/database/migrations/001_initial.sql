CREATE TABLE users (
    id text PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL UNIQUE
);

CREATE TABLE rooms (
    id text PRIMARY KEY,
    name text NOT NULL,
    capacity integer NOT NULL CHECK (capacity > 0),
    amenities jsonb NOT NULL DEFAULT '[]'::jsonb,
    description text NOT NULL
);

CREATE TABLE bookings (
    id text PRIMARY KEY,
    room_id text NOT NULL REFERENCES rooms(id),
    organizer_id text NOT NULL REFERENCES users(id),
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    purpose text NOT NULL,
    CHECK (starts_at < ends_at)
);

CREATE INDEX bookings_room_time_idx ON bookings(room_id, starts_at, ends_at);

CREATE TABLE calendar_events (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES users(id),
    title text NOT NULL,
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    CHECK (starts_at < ends_at)
);

CREATE INDEX calendar_events_user_time_idx ON calendar_events(user_id, starts_at, ends_at);

CREATE TABLE sessions (
    token text PRIMARY KEY,
    user_id text NOT NULL REFERENCES users(id),
    expires_at timestamptz NOT NULL
);
