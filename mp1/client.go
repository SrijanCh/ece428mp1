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
    	// ips.PushBack(line[:len(line)-1])
    	ips.PushBack(line)
    	// fmt.Printf("%s, abc", line[:len(line)-1])
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
		if len(s_addr) - 1 < 0{
			continue
		}
		rval := rgrep(s_addr[:len(s_addr)-1], data);
		if rval == 1 {
			fmt.Printf("Failed to connect to %s\n", addr.Value)
		}
		if rval == 2 {
			fmt.Printf("RPC failed at %s\n", addr.Value)
		}
	}
}



func rgrep(addr, arg string) int{
	// fmt.Printf("Setting connection to %s\n", addr + ":3074")
	client, err := rpc.DialHTTP("tcp", addr + ":3074")
	if err != nil {
		// fmt.Printf(err.Error())
		// fmt.Printf("\n")
		return 1
	}
	// Synchronous call
	var args = querier.Args{Data: arg, Filepath: "machine.i.log"}
	var reply string
	err = client.Call("Querier.Grep", args, &reply)
	if err != nil {
		return 2
	}
	fmt.Printf("%s\n", reply)
	return 0
}