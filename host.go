package main
import(
		"fmt"
		"log"
	   "net/rpc"
	   "net/http"
	   "net"
	   "server"
	   )
func main(){
	arith := new(server.Arith)
	rpc.Register(arith)
	rpc.HandleHTTP()
	fmt.Printf("Start listening: \n");
	l, e := net.Listen("tcp", "127.0.0.1:3074")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		conn.Close()
	}
	fmt.Printf("Done. \n");
	go http.Serve(l, nil)
}