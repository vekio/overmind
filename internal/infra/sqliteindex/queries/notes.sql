-- name: UpsertNote :exec
INSERT INTO notes (id, kind, created_at, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    kind = excluded.kind,
    created_at = excluded.created_at,
    updated_at = excluded.updated_at;

-- name: DeleteNoteAttributes :exec
DELETE FROM note_attributes
WHERE note_id = ?;

-- name: InsertNoteAttribute :exec
INSERT INTO note_attributes (note_id, name, value)
VALUES (?, ?, ?);

-- name: JournalExists :one
SELECT EXISTS (
    SELECT 1
    FROM notes
    JOIN note_attributes ON note_attributes.note_id = notes.id
    WHERE notes.kind = 'journal'
      AND note_attributes.name = 'date'
      AND note_attributes.value = ?
);

-- name: DeleteNoteTags :exec
DELETE FROM note_tags
WHERE note_id = ?;

-- name: InsertNoteTag :exec
INSERT INTO note_tags (note_id, tag, position)
VALUES (?, ?, ?);

-- name: ListNotes :many
SELECT id, kind, created_at, updated_at
FROM notes
ORDER BY updated_at DESC, id DESC;

-- name: ListNoteAttributes :many
SELECT note_id, name, value
FROM note_attributes
ORDER BY note_id, name;

-- name: ListNoteTags :many
SELECT note_id, tag, position
FROM note_tags
ORDER BY note_id, position;
