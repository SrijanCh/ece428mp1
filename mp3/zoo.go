package main
import(
	// "sdfsrpc"
	"zookeeper"
	"log"
	"net/rpc"
	"net/http"
	"net"
	"fmt"
)

func main(){
	zookeeper := new(zookeeper.Zookeeper) //Creates a new Zookeeper object to handle the RPCs for this server
	rpc.Register(zookeeper) //Registers the Zookeeper as our handler
	rpc.HandleHTTP() //HTTP format requests
	fmt.Printf("Zookeeper is up for business! \n"); //Notify user
	l, e := net.Listen("tcp", ":3074") //Listen to requests on port 3074
	if e != nil {
		//Error handling
		log.Fatal("listen error:", e)
	}
	fmt.Printf("Serving on port %d\n", 3074); //Notify that we're up
	http.Serve(l, nil) //Serve
}
