package main
import(
		"fmt"
	    "net/rpc"
	    "os"
		"querier"
    	"bufio"
    	"container/list"
	  )

func main(){
	ips := list.New()

	file, err := os.Open("../iptables.txt")
	defer file.Close()

	if err != nil {
    	fmt.Println("Error opening file")
    	return	
	}

	reader := bufio.NewReader(file)
	var line string
	for {
    	line, err = reader.ReadString('\n')
    	ips.PushBack(line)
    	if err != nil {
        	break
    	}
    }

    for addr := ips.Front(); addr != nil; addr = addr.Next() {
		var data = ""
		if len(os.Args) > 1 {
			data = os.Args[1]
		}
		var s_addr string = fmt.Sprintf("%v", addr.Value)
		rval := rgrep(s_addr, data);
		if rval == 1 {
			fmt.Printf("Failed to connect to %s\n", addr.Value)
		}
		if rval == 2 {
			fmt.Printf("RPC failed at %s\n", addr.Value)
		}
	}
}



func rgrep(addr, arg string) int{
	client, err := rpc.DialHTTP("tcp", addr + ":3074")
	if err != nil {
		fmt.Printf(err.Error())
		return 1
	}
	// Synchronous call
	var args = querier.Args{Data: arg, Filepath: "~/machine.i.log"}
	var reply string
	err = client.Call("Querier.Grep", args, &reply)
	if err != nil {
		return 2
	}
	fmt.Printf(reply)
	return 0
}