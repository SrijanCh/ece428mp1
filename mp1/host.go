package main
import(
		"fmt"
		"log"
	   	"net/rpc"
	   	"net/http"
	   	"net"
	   	"querier"
	   )

//Running this program allows us to host a server on our machine to handle grep queries
func main(){
	querier := new(querier.Querier) //Creates a new Querier object to handle the RPCs for this server
	rpc.Register(querier) //Registers the Querier as our handler
	rpc.HandleHTTP() //HTTP format requests
	fmt.Printf("Start listening: \n"); //Notify user
	l, e := net.Listen("tcp", ":3074") //Listen to requests on port 3074
	if e != nil {
		//Error handling
		log.Fatal("listen error:", e)
	}
	fmt.Printf("Serving RPC server on port %d\n", 3074); //Notify that we're up
	http.Serve(l, nil) //Serve
}