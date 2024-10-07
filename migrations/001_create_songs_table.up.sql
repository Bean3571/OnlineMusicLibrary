CREATE TABLE IF NOT EXISTS songs (
    id SERIAL PRIMARY KEY,
    group VARCHAR(255) NOT NULL,
    song VARCHAR(255) NOT NULL,
    release_date VARCHAR(255),
    text TEXT,
    link VARCHAR(255)
);
