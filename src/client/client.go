package main
import(
		"fmt"
	    "net/rpc"
	    "log"
		// "querier"
		// "strings"
	  )

type Args struct {
	Data, Filepath string
}

func main(){
	client, err := rpc.DialHTTP("tcp", "localhost" + ":1234")
	if err != nil {log.Fatal("dialing:", err)}
	// Synchronous call
	args := Args{Data: "He", Filepath: "machine.i.log"}
	var reply string
	fmt.Printf("%T, %T\n", args, reply)
	err = client.Call("Querier.Grep", args, &reply)
	if err != nil {
		fmt.Printf("WE FUCKED UP\n")
		log.Fatal("query error: ", err)
	}
	fmt.Printf("grep result: %s\n", reply)
}
