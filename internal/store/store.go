package store

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"time"
)

type DB struct {
	db      *sql.DB
	dataDir string
}
type File struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Bucket      string `json:"bucket"`
	Tags        string `json:"tags"`
	CreatedAt   string `json:"created_at"`
}
type Bucket struct {
	Name      string `json:"name"`
	FileCount int    `json:"file_count"`
	TotalSize int64  `json:"total_size"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(d, "files"), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "silo.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS files(id TEXT PRIMARY KEY,name TEXT NOT NULL,size INTEGER DEFAULT 0,content_type TEXT DEFAULT 'application/octet-stream',bucket TEXT DEFAULT 'default',tags TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')));CREATE TABLE IF NOT EXISTS extras(resource TEXT NOT NULL,record_id TEXT NOT NULL,data TEXT NOT NULL DEFAULT '{}',PRIMARY KEY(resource, record_id));`)
	return &DB{db: db, dataDir: d}, nil
}
func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }
func (d *DB) SaveFile(f *File, data []byte) error {
	f.ID = genID()
	f.CreatedAt = now()
	f.Size = int64(len(data))
	if f.Bucket == "" {
		f.Bucket = "default"
	}
	os.MkdirAll(filepath.Join(d.dataDir, "files", f.Bucket), 0755)
	if err := os.WriteFile(filepath.Join(d.dataDir, "files", f.Bucket, f.ID), data, 0644); err != nil {
		return err
	}
	_, err := d.db.Exec(`INSERT INTO files(id,name,size,content_type,bucket,tags,created_at)VALUES(?,?,?,?,?,?,?)`, f.ID, f.Name, f.Size, f.ContentType, f.Bucket, f.Tags, f.CreatedAt)
	return err
}
func (d *DB) GetFile(id string) *File {
	var f File
	if d.db.QueryRow(`SELECT id,name,size,content_type,bucket,tags,created_at FROM files WHERE id=?`, id).Scan(&f.ID, &f.Name, &f.Size, &f.ContentType, &f.Bucket, &f.Tags, &f.CreatedAt) != nil {
		return nil
	}
	return &f
}
func (d *DB) ReadFile(id string) ([]byte, error) {
	f := d.GetFile(id)
	if f == nil {
		return nil, fmt.Errorf("not found")
	}
	return os.ReadFile(filepath.Join(d.dataDir, "files", f.Bucket, f.ID))
}
func (d *DB) List(bucket string) []File {
	q := `SELECT id,name,size,content_type,bucket,tags,created_at FROM files`
	args := []any{}
	if bucket != "" {
		q += ` WHERE bucket=?`
		args = append(args, bucket)
	}
	q += ` ORDER BY created_at DESC`
	rows, _ := d.db.Query(q, args...)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []File
	for rows.Next() {
		var f File
		rows.Scan(&f.ID, &f.Name, &f.Size, &f.ContentType, &f.Bucket, &f.Tags, &f.CreatedAt)
		o = append(o, f)
	}
	return o
}
func (d *DB) Delete(id string) error {
	f := d.GetFile(id)
	if f != nil {
		os.Remove(filepath.Join(d.dataDir, "files", f.Bucket, f.ID))
	}
	_, err := d.db.Exec(`DELETE FROM files WHERE id=?`, id)
	return err
}
func (d *DB) ListBuckets() []Bucket {
	rows, _ := d.db.Query(`SELECT bucket,COUNT(*),COALESCE(SUM(size),0) FROM files GROUP BY bucket ORDER BY bucket`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Bucket
	for rows.Next() {
		var b Bucket
		rows.Scan(&b.Name, &b.FileCount, &b.TotalSize)
		o = append(o, b)
	}
	return o
}
func (d *DB) Stats() map[string]any {
	var files int
	var size int64
	d.db.QueryRow(`SELECT COUNT(*),COALESCE(SUM(size),0) FROM files`).Scan(&files, &size)
	buckets := d.ListBuckets()
	return map[string]any{"files": files, "total_size": size, "buckets": len(buckets)}
}

// ─── Extras: generic key-value storage for personalization custom fields ───

func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(
		`SELECT data FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(
		`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?)
		 ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`,
		resource, recordID, data,
	)
	return err
}

func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(
		`DELETE FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	)
	return err
}

func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(
		`SELECT record_id, data FROM extras WHERE resource=?`,
		resource,
	)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
