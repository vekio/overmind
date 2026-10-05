-- name: UpsertPerson :exec
INSERT INTO persons (note_id, name) VALUES (?, ?)
ON CONFLICT(note_id) DO UPDATE SET name=excluded.name;

-- name: EnsureGroup :exec
INSERT INTO groups (name) VALUES (?) ON CONFLICT(name) DO NOTHING;

-- name: DeletePersonGroups :exec
DELETE FROM person_groups WHERE person_id = ?;

-- name: InsertPersonGroup :exec
INSERT INTO person_groups (person_id, group_name, position) VALUES (?, ?, ?);

-- name: PersonByID :one
SELECT n.id, n.type, n.path, n.created_at, n.updated_at, p.name
FROM notes n JOIN persons p ON p.note_id = n.id WHERE n.id = ? AND n.type = 'person';

-- name: GroupsByPersonID :many
SELECT group_name FROM person_groups WHERE person_id = ? ORDER BY position;

-- name: FindGroups :many
SELECT DISTINCT group_name FROM person_groups ORDER BY group_name;
