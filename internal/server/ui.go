package server
import "net/http"
func(s *Server)dashboard(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html");w.Write([]byte(dashHTML))}
const dashHTML=`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Silo</title>
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}
.main{padding:1.5rem;max-width:900px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:.6rem;margin-bottom:1.2rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem;text-align:center}.st-v{font-size:1.2rem}.st-l{font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.1rem}
.upload{background:var(--bg2);border:2px dashed var(--bg3);padding:1.5rem;text-align:center;margin-bottom:1.2rem;cursor:pointer;transition:border-color .2s}
.upload:hover{border-color:var(--leather)}
.upload input{display:none}
.bucket-bar{display:flex;gap:.3rem;margin-bottom:1rem;flex-wrap:wrap}
.bucket-btn{font-size:.6rem;padding:.2rem .5rem;border:1px solid var(--bg3);background:var(--bg);color:var(--cm);cursor:pointer}.bucket-btn:hover{border-color:var(--leather)}.bucket-btn.active{border-color:var(--gold);color:var(--gold)}
.file{display:flex;justify-content:space-between;align-items:center;padding:.5rem .8rem;border-bottom:1px solid var(--bg3);font-size:.75rem}
.file:hover{background:var(--bg2)}
.file-name{color:var(--cream)}.file-meta{font-size:.6rem;color:var(--cm);display:flex;gap:.8rem}
.btn{font-size:.6rem;padding:.2rem .5rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1>SILO</h1></div>
<div class="main">
<div class="stats" id="stats"></div>
<div class="upload" onclick="document.getElementById('fileInput').click()"><input type="file" id="fileInput" multiple onchange="upload(this.files)">Drop files here or click to upload<div style="font-size:.6rem;color:var(--cm);margin-top:.3rem">Files stored on disk, metadata in SQLite</div></div>
<div class="bucket-bar" id="buckets"></div>
<div id="files"></div>
</div>
<script>
const A='/api';let files=[],buckets=[],curBucket='';
async function load(){const[f,b,s]=await Promise.all([fetch(A+'/files'+(curBucket?'?bucket='+encodeURIComponent(curBucket):'')).then(r=>r.json()),fetch(A+'/buckets').then(r=>r.json()),fetch(A+'/stats').then(r=>r.json())]);
files=f.files||[];buckets=b.buckets||[];
document.getElementById('stats').innerHTML='<div class="st"><div class="st-v">'+s.files+'</div><div class="st-l">Files</div></div><div class="st"><div class="st-v">'+fmtSize(s.total_size)+'</div><div class="st-l">Total Size</div></div><div class="st"><div class="st-v">'+s.buckets+'</div><div class="st-l">Buckets</div></div>';
let bh='<button class="bucket-btn'+(curBucket===''?' active':'')+'" onclick="setBucket(\'\')">All</button>';
buckets.forEach(b=>{bh+='<button class="bucket-btn'+(curBucket===b.name?' active':'')+'" onclick="setBucket(\''+b.name+'\')">'+esc(b.name)+' ('+b.file_count+')</button>';});
document.getElementById('buckets').innerHTML=bh;render();}
function setBucket(b){curBucket=b;load();}
function render(){if(!files.length){document.getElementById('files').innerHTML='<div class="empty">No files uploaded yet.</div>';return;}
let h='';files.forEach(f=>{h+='<div class="file"><div><span class="file-name">'+esc(f.name)+'</span></div><div class="file-meta"><span>'+fmtSize(f.size)+'</span><span>'+f.content_type+'</span><span>'+f.bucket+'</span><a href="'+A+'/files/'+f.id+'/download" class="btn">Download</a><button class="btn" onclick="del(\''+f.id+'\')" style="color:var(--cm)">✕</button></div></div>';});
document.getElementById('files').innerHTML=h;}
async function upload(fileList){for(const file of fileList){const fd=new FormData();fd.append('file',file);fd.append('bucket',curBucket||'default');await fetch(A+'/files',{method:'POST',body:fd});}load();}
async function del(id){if(confirm('Delete?')){await fetch(A+'/files/'+id,{method:'DELETE'});load();}}
function fmtSize(b){if(b>=1073741824)return(b/1073741824).toFixed(1)+'GB';if(b>=1048576)return(b/1048576).toFixed(1)+'MB';if(b>=1024)return(b/1024).toFixed(1)+'KB';return b+'B';}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
