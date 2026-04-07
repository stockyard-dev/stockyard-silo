package server

import (
	"encoding/json"
	"github.com/stockyard-dev/stockyard-silo/internal/store"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	db      *store.DB
	mux     *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.mux.HandleFunc("GET /api/files", s.listFiles)
	s.mux.HandleFunc("POST /api/files", s.uploadFile)
	s.mux.HandleFunc("GET /api/files/{id}", s.getFile)
	s.mux.HandleFunc("GET /api/files/{id}/download", s.downloadFile)
	s.mux.HandleFunc("DELETE /api/files/{id}", s.deleteFile)
	s.mux.HandleFunc("GET /api/buckets", s.listBuckets)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/silo/"})
	})
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)
	return s
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	json.NewEncoder(w).Encode(v)
}
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", 302)
}
func (s *Server) listFiles(w http.ResponseWriter, r *http.Request) {
	bucket := r.URL.Query().Get("bucket")
	files := s.db.List(bucket)
	if files == nil {
		files = []store.File{}
	}
	wj(w, 200, map[string]any{"files": files})
}
func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		we(w, 400, "file required")
		return
	}
	defer file.Close()
	data, _ := io.ReadAll(file)
	f := &store.File{Name: header.Filename, ContentType: header.Header.Get("Content-Type"), Bucket: r.FormValue("bucket"), Tags: r.FormValue("tags")}
	if f.ContentType == "" {
		f.ContentType = "application/octet-stream"
	}
	s.db.SaveFile(f, data)
	wj(w, 201, f)
}
func (s *Server) getFile(w http.ResponseWriter, r *http.Request) {
	f := s.db.GetFile(r.PathValue("id"))
	if f == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, f)
}
func (s *Server) downloadFile(w http.ResponseWriter, r *http.Request) {
	f := s.db.GetFile(r.PathValue("id"))
	if f == nil {
		we(w, 404, "not found")
		return
	}
	data, err := s.db.ReadFile(r.PathValue("id"))
	if err != nil {
		we(w, 500, err.Error())
		return
	}
	w.Header().Set("Content-Type", f.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+f.Name+"\"")
	w.Write(data)
}
func (s *Server) deleteFile(w http.ResponseWriter, r *http.Request) {
	s.db.Delete(r.PathValue("id"))
	wj(w, 200, map[string]string{"status": "deleted"})
}
func (s *Server) listBuckets(w http.ResponseWriter, r *http.Request) {
	buckets := s.db.ListBuckets()
	if buckets == nil {
		buckets = []store.Bucket{}
	}
	wj(w, 200, map[string]any{"buckets": buckets})
}
func (s *Server) stats(w http.ResponseWriter, r *http.Request) { wj(w, 200, s.db.Stats()) }
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats()
	wj(w, 200, map[string]any{"service": "silo", "status": "ok", "files": st["files"], "total_size": st["total_size"]})
}

// ─── personalization (auto-added) ──────────────────────────────────

func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("%s: warning: could not parse config.json: %v", "silo", err)
		return
	}
	s.pCfg = cfg
	log.Printf("%s: loaded personalization from %s", "silo", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"read body"}`, 400)
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		http.Error(w, `{"error":"invalid json"}`, 400)
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		http.Error(w, `{"error":"save failed"}`, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":"saved"}`))
}
