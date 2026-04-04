package server
import ("encoding/json";"io";"net/http";"github.com/stockyard-dev/stockyard-silo/internal/store")
type Server struct{db *store.DB;mux *http.ServeMux;limits Limits}
func New(db *store.DB,limits Limits)*Server{s:=&Server{db:db,mux:http.NewServeMux(),limits:limits}
s.mux.HandleFunc("GET /api/files",s.listFiles)
s.mux.HandleFunc("POST /api/files",s.uploadFile)
s.mux.HandleFunc("GET /api/files/{id}",s.getFile)
s.mux.HandleFunc("GET /api/files/{id}/download",s.downloadFile)
s.mux.HandleFunc("DELETE /api/files/{id}",s.deleteFile)
s.mux.HandleFunc("GET /api/buckets",s.listBuckets)
s.mux.HandleFunc("GET /api/stats",s.stats)
s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /api/tier",func(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"tier":s.limits.Tier,"upgrade_url":"https://stockyard.dev/silo/"})})
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root)
return s}
func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)listFiles(w http.ResponseWriter,r *http.Request){bucket:=r.URL.Query().Get("bucket");files:=s.db.List(bucket);if files==nil{files=[]store.File{}};wj(w,200,map[string]any{"files":files})}
func(s *Server)uploadFile(w http.ResponseWriter,r *http.Request){r.ParseMultipartForm(32<<20);file,header,err:=r.FormFile("file");if err!=nil{we(w,400,"file required");return};defer file.Close()
data,_:=io.ReadAll(file);f:=&store.File{Name:header.Filename,ContentType:header.Header.Get("Content-Type"),Bucket:r.FormValue("bucket"),Tags:r.FormValue("tags")}
if f.ContentType==""{f.ContentType="application/octet-stream"}
s.db.SaveFile(f,data);wj(w,201,f)}
func(s *Server)getFile(w http.ResponseWriter,r *http.Request){f:=s.db.GetFile(r.PathValue("id"));if f==nil{we(w,404,"not found");return};wj(w,200,f)}
func(s *Server)downloadFile(w http.ResponseWriter,r *http.Request){f:=s.db.GetFile(r.PathValue("id"));if f==nil{we(w,404,"not found");return}
data,err:=s.db.ReadFile(r.PathValue("id"));if err!=nil{we(w,500,err.Error());return}
w.Header().Set("Content-Type",f.ContentType);w.Header().Set("Content-Disposition","attachment; filename=\""+f.Name+"\"");w.Write(data)}
func(s *Server)deleteFile(w http.ResponseWriter,r *http.Request){s.db.Delete(r.PathValue("id"));wj(w,200,map[string]string{"status":"deleted"})}
func(s *Server)listBuckets(w http.ResponseWriter,r *http.Request){buckets:=s.db.ListBuckets();if buckets==nil{buckets=[]store.Bucket{}};wj(w,200,map[string]any{"buckets":buckets})}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){st:=s.db.Stats();wj(w,200,map[string]any{"service":"silo","status":"ok","files":st["files"],"total_size":st["total_size"]})}
