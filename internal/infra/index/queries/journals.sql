-- name: UpsertJournal :exec
INSERT INTO journals(note_id,date) VALUES(?,?)
ON CONFLICT(note_id) DO UPDATE SET date=excluded.date;

-- name: JournalExists :one
SELECT EXISTS(SELECT 1 FROM journals WHERE date=?);
