package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-silo/internal/server";"github.com/stockyard-dev/stockyard-silo/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="9803"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./silo-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("silo: %v",err)};defer db.Close();srv:=server.New(db)
fmt.Printf("\n  Silo — Self-hosted data isolation and partitioning layer\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n\n",port,port)
log.Printf("silo: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
