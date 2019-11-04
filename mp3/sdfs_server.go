package main
import(
	"sdfsrpc"
	"log"
	"net/rpc"
	"net/http"
	"net"
	"fmt"
)
const node_portnum = "3074"
func main(){
	sdfsrpc := new(sdfsrpc.Sdfsrpc) //Creates a new Querier object to handle the RPCs for this server
	rpc.Register(sdfsrpc) //Registers the Querier as our handler
	rpc.HandleHTTP() //HTTP format requests
	fmt.Printf("Sdfsrpc server start listening: \n"); //Notify user
	l, e := net.Listen("tcp", ":" + node_portnum)//3074") //Listen to requests on port 3074
	if e != nil {
		//Error handling
		log.Fatal("listen error:", e)
	}
	fmt.Printf("Serving RPC server on port %d\n", 3074); //Notify that we're up
	http.Serve(l, nil) //Serve
}
