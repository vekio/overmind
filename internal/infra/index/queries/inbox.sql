-- name: UpsertInbox :exec
INSERT INTO inbox (note_id, content) VALUES (?, ?)
ON CONFLICT(note_id) DO UPDATE SET content=excluded.content;
