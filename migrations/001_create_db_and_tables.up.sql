CREATE TABLE IF NOT EXISTS musicGroups (
    id_group        SERIAL PRIMARY KEY,
    groupName         VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS songs (
    id_song         SERIAL PRIMARY KEY,
    id_group        INT NOT NULL,
    song            VARCHAR(255) NOT NULL,
    release_date    DATE,
    lyrics          TEXT,
    link            VARCHAR(255),
    CONSTRAINT fk_id_group FOREIGN KEY (id_group) REFERENCES musicGroups (id_group) ON DELETE CASCADE
);
