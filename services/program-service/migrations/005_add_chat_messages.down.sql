DROP INDEX IF EXISTS idx_chat_messages_created_at;
DROP INDEX IF EXISTS idx_chat_messages_conversation_id;
DROP INDEX IF EXISTS idx_chat_conversations_user_id;

DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_conversations;
