package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Partition struct {
	ID string `json:"id"`
	Name string `json:"name"`
	TenantID string `json:"tenant_id"`
	DataSource string `json:"data_source"`
	RecordCount int `json:"record_count"`
	SizeBytes int `json:"size_bytes"`
	Status string `json:"status"`
	IsolationLevel string `json:"isolation_level"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"silo.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS partitions(id TEXT PRIMARY KEY,name TEXT NOT NULL,tenant_id TEXT DEFAULT '',data_source TEXT DEFAULT '',record_count INTEGER DEFAULT 0,size_bytes INTEGER DEFAULT 0,status TEXT DEFAULT 'active',isolation_level TEXT DEFAULT 'strict',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Partition)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO partitions(id,name,tenant_id,data_source,record_count,size_bytes,status,isolation_level,created_at)VALUES(?,?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.TenantID,e.DataSource,e.RecordCount,e.SizeBytes,e.Status,e.IsolationLevel,e.CreatedAt);return err}
func(d *DB)Get(id string)*Partition{var e Partition;if d.db.QueryRow(`SELECT id,name,tenant_id,data_source,record_count,size_bytes,status,isolation_level,created_at FROM partitions WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.TenantID,&e.DataSource,&e.RecordCount,&e.SizeBytes,&e.Status,&e.IsolationLevel,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Partition{rows,_:=d.db.Query(`SELECT id,name,tenant_id,data_source,record_count,size_bytes,status,isolation_level,created_at FROM partitions ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Partition;for rows.Next(){var e Partition;rows.Scan(&e.ID,&e.Name,&e.TenantID,&e.DataSource,&e.RecordCount,&e.SizeBytes,&e.Status,&e.IsolationLevel,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Update(e *Partition)error{_,err:=d.db.Exec(`UPDATE partitions SET name=?,tenant_id=?,data_source=?,record_count=?,size_bytes=?,status=?,isolation_level=? WHERE id=?`,e.Name,e.TenantID,e.DataSource,e.RecordCount,e.SizeBytes,e.Status,e.IsolationLevel,e.ID);return err}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM partitions WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM partitions`).Scan(&n);return n}

func(d *DB)Search(q string, filters map[string]string)[]Partition{
    where:="1=1"
    args:=[]any{}
    if q!=""{
        where+=" AND (name LIKE ?)"
        args=append(args,"%"+q+"%");
    }
    if v,ok:=filters["status"];ok&&v!=""{where+=" AND status=?";args=append(args,v)}
    rows,_:=d.db.Query(`SELECT id,name,tenant_id,data_source,record_count,size_bytes,status,isolation_level,created_at FROM partitions WHERE `+where+` ORDER BY created_at DESC`,args...)
    if rows==nil{return nil};defer rows.Close()
    var o []Partition;for rows.Next(){var e Partition;rows.Scan(&e.ID,&e.Name,&e.TenantID,&e.DataSource,&e.RecordCount,&e.SizeBytes,&e.Status,&e.IsolationLevel,&e.CreatedAt);o=append(o,e)};return o
}

func(d *DB)Stats()map[string]any{
    m:=map[string]any{"total":d.Count()}
    rows,_:=d.db.Query(`SELECT status,COUNT(*) FROM partitions GROUP BY status`)
    if rows!=nil{defer rows.Close();by:=map[string]int{};for rows.Next(){var s string;var c int;rows.Scan(&s,&c);by[s]=c};m["by_status"]=by}
    return m
}
