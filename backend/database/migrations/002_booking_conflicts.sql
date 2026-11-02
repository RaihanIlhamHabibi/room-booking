CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE bookings
    ADD CONSTRAINT bookings_no_room_overlap
    EXCLUDE USING gist (
        room_id WITH =,
        tstzrange(starts_at, ends_at, '[)') WITH &&
    );

ALTER TABLE bookings
    ADD CONSTRAINT bookings_purpose_length
    CHECK (char_length(btrim(purpose)) BETWEEN 1 AND 200);
