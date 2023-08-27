CREATE TABLE segments (
    id SERIAL PRIMARY KEY NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE users_segments (
    id SERIAL PRIMARY KEY NOT NULL UNIQUE,
    segment_id INT NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY (segment_id) REFERENCES segments(id),

    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    removed_at TIMESTAMP, -- if this column is not null, this record was removed, either manually or by deleting associated segment
    expires_at TIMESTAMP -- if this column is not null, this record has an expiration date. if the timestamp is in the past, it has already expired
);