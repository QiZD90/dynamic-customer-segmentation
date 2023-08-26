CREATE TABLE segments (
    id SERIAL PRIMARY KEY NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE users_segments (
    id SERIAL PRIMARY KEY NOT NULL UNIQUE,
    segment_id INT NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY (segment_id) REFERENCES segments(id),

    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    removed_at TIMESTAMP,
    expires_at TIMESTAMP
);