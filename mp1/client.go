package main
import(
		"fmt"
	    "net/rpc"
	    // "log"
	    "os"
		"querier"
		// "strings"
	  )

// type Args struct {
// 	Data, Filepath string
// }

func main(){
	// client, err := rpc.DialHTTP("tcp", "localhost" + ":1234")
	// if err != nil {log.Fatal("dialing:", err)}
	// // Synchronous call
	// var args = querier.Args{Data: "", Filepath: "machine.i.log"}
	// if len(os.Args) > 1 {
	// 	args = querier.Args{Data: os.Args[1], Filepath: "machine.i.log"}
	// }
	// var reply string
	// // fmt.Printf("%T, %T\n", args, reply)
	// err = client.Call("Querier.Grep", args, &reply)
	// if err != nil {
	// 	// fmt.Printf("WE FUCKED UP\n")
	// 	log.Fatal("query error: ", err)
	// }
	// name, err := os.Hostname()
	// if err != nil {
 	//     	fmt.Printf("Oops: %v\n", err)
 	//    	return
	// }
	// fmt.Printf("%s reports:\n %s\n", name, reply)\
	var data = ""
	var addr = "localhost"
	if len(os.Args) > 1 {
		data = os.Args[1]
	}
	
	rval := rgrep(addr, data);
	if rval == 1 {
		fmt.Printf("Failed to connect to %s\n", addr)
	}
	if rval == 2 {
		fmt.Printf("RPC failed at %s\n", addr)
	}
}



func rgrep(addr, arg string) int{
	client, err := rpc.DialHTTP("tcp", addr + ":1234")
	if err != nil {
		return 1
	}
	// Synchronous call
	var args = querier.Args{Data: arg, Filepath: "machine.i.log"}
	var reply string
	err = client.Call("Querier.Grep", args, &reply)
	if err != nil {
		return 2
	}
	fmt.Printf(reply)
	return 0
}