CREATE INDEX IF NOT EXISTS idx_chat_messages_room_created_at ON chat_messages(room_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_active ON chat_messages(room_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_room_presence_room_online ON room_presence(room_id, online);
