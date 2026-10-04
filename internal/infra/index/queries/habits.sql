-- name: UpsertNote :exec
INSERT INTO notes (id, type, path, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET type=excluded.type, path=excluded.path,
 created_at=excluded.created_at, updated_at=excluded.updated_at;

-- name: UpsertHabit :exec
INSERT INTO habits (note_id, title, target_amount, unit, period) VALUES (?, ?, ?, ?, ?)
ON CONFLICT(note_id) DO UPDATE SET title=excluded.title, target_amount=excluded.target_amount, unit=excluded.unit, period=excluded.period;

-- name: DeleteNoteTags :exec
DELETE FROM note_tags WHERE note_id = ?;

-- name: EnsureTag :exec
INSERT INTO tags (name) VALUES (?) ON CONFLICT(name) DO NOTHING;

-- name: InsertNoteTag :exec
INSERT INTO note_tags (note_id, tag, position) VALUES (?, ?, ?);

-- name: HabitByID :one
SELECT n.id, n.type, n.path, h.title, n.created_at, n.updated_at,
 h.target_amount, h.unit, h.period
FROM notes n JOIN habits h ON h.note_id = n.id WHERE n.id = ? AND n.type = 'habit';

-- name: TagsByNoteID :many
SELECT tag FROM note_tags WHERE note_id = ? ORDER BY position;

-- name: DeleteNote :exec
DELETE FROM notes WHERE id = ?;
