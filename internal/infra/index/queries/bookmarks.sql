-- name: UpsertBookmark :exec
INSERT INTO bookmarks (note_id, url) VALUES (?, ?)
ON CONFLICT(note_id) DO UPDATE SET url=excluded.url;
