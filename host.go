package main
import(
		"fmt"
		"log"
	   "net/rpc"
	   "net/http"
	   "net"
	   "rpctest"
	   )
func main(){
	arith := new(rpctest.Arith)
	rpc.Register(arith)
	rpc.HandleHTTP()
	fmt.Printf("Start listening: \n");
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	// for {
	// 	conn, err := l.Accept()
	// 	if err != nil {
	// 		log.Fatal("accept error:", err)
	// 	} else{
	// 		fmt.Printf("GOT ONE\n")
	// 	}
	// 	conn.LocalAddr()
	// 	conn.Close();
	// 	// if a > 0 {
	// 	// 	fmt.Printf("stfu")
	// 	// }
	// }
	
	fmt.Printf("Serving RPC server on port %d", 1234);
	http.Serve(l, nil)
}