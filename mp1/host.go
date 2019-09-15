package main
import(
		"fmt"
		"log"
	   	"net/rpc"
	   	"net/http"
	   	"net"
	   	"querier"
	   )
func main(){
	querier := new(querier.Querier)
	rpc.Register(querier)
	rpc.HandleHTTP()
	fmt.Printf("Start listening: \n");
	l, e := net.Listen("tcp", ":3074")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Printf("Serving RPC server on port %d\n", 3074);
	http.Serve(l, nil)
}