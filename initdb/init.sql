CREATE TABLE users (
    id UUID PRIMARY KEY,
    email text not null,
    refresh_token_hash text
);