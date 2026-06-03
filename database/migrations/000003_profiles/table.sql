CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL DEFAULT '',
    bio TEXT NOT NULL DEFAULT '',
    headline TEXT NOT NULL DEFAULT '',
    education TEXT NOT NULL DEFAULT '',
    occupation TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    website_url TEXT NOT NULL DEFAULT '',
    avatar_object_id UUID,
    cover_object_id UUID,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
