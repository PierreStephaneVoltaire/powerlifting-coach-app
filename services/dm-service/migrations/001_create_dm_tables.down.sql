DROP TRIGGER IF EXISTS update_messages_updated_at ON messages;
DROP TRIGGER IF EXISTS update_conversations_updated_at ON conversations;

DROP INDEX IF EXISTS idx_idempotency_keys_event_type;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_sender_id;
DROP INDEX IF EXISTS idx_messages_conversation_id;
DROP INDEX IF EXISTS idx_conversations_last_message_at;
DROP INDEX IF EXISTS idx_conversations_participant_2;
DROP INDEX IF EXISTS idx_conversations_participant_1;

DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
