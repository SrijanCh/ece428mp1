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

    var i int = 1
    for addr := ips.Front(); addr != nil; addr = addr.Next() {
		var data string = ""

		for j := 1; j < len(os.Args); j++{
			data += os.Args[j] + " "
		}

	// fmt.Printf("grep" + data + "standin\n")
		// if len(os.Args) > 1 {
		// 	data = os.Args[1]
		// }
		
		var s_addr string = fmt.Sprintf("%v", addr.Value)
		if len(s_addr) - 1 < 0{
			continue
		}

		var filep string = fmt.Sprintf("vm%d.log", i)

		rval := homegrep(s_addr[:len(s_addr)-1], data, filep);
		if rval == 1 {
			fmt.Printf("Failed to connect to %s\n", addr.Value)
		}
		if rval == 2 {
			fmt.Printf("RPC failed at %s\n", addr.Value)
		}

		i++
	}
}

func loggrep(addr, arg string) int{
	client, err := rpc.DialHTTP("tcp", addr + ":3074")
	if err != nil {
		// fmt.Printf(err.Error())
		// fmt.Printf("\n")
		return 1
	}
	// Synchronous call
	var args = querier.Args{Data: arg, Filepath: "/home/srijanc2/machine.i.log"}
	var reply string
	err = client.Call("Querier.Grep", args, &reply)
	if err != nil {
		return 2
	}
	fmt.Printf("%s\n", reply)
	return 0
}

func rgrep(addr, arg, filepath string) int{
	// fmt.Printf("Setting connection to %s\n", addr + ":3074")
	client, err := rpc.DialHTTP("tcp", addr + ":3074")
	if err != nil {
		// fmt.Printf(err.Error())
		// fmt.Printf("\n")
		return 1
	}
	// Synchronous call
	var args = querier.Args{Data: arg, Filepath: filepath}
	var reply string
	err = client.Call("Querier.Grep", args, &reply)
	if err != nil {
		return 2
	}
	fmt.Printf("%s\n", reply)
	return 0
}

func homegrep(addr, arg, filepath string) int{
	// fmt.Printf("Setting connection to %s\n", addr + ":3074")
	client, err := rpc.DialHTTP("tcp", addr + ":3074")
	if err != nil {
		// fmt.Printf(err.Error())
		// fmt.Printf("\n")
		return 1
	}
	// Synchronous call
	var args = querier.Args{Data: arg, Filepath: "/home/srijanc2/" + filepath}
	var reply string
	err = client.Call("Querier.Grep", args, &reply)
	if err != nil {
		return 2
	}
	fmt.Printf("%s\n", reply)
	return 0
}


func joey_loggrep(addr, arg string) int{
	// fmt.Printf("Setting connection to %s\n", addr + ":3074")
	client, err := rpc.DialHTTP("tcp", addr + ":3074")
	if err != nil {
		// fmt.Printf(err.Error())
		// fmt.Printf("\n")
		return 1
	}
	// Synchronous call
	var args = querier.Args{Data: arg, Filepath: "/home/jbahary2/machine.i.log"}
	var reply string
	err = client.Call("Querier.Grep", args, &reply)
	if err != nil {
		return 2
	}
	fmt.Printf("%s\n", reply)
	return 0
}


func joey_homegrep(addr, arg, filepath string) int{
	// fmt.Printf("Setting connection to %s\n", addr + ":3074")
	client, err := rpc.DialHTTP("tcp", addr + ":3074")
	if err != nil {
		// fmt.Printf(err.Error())
		// fmt.Printf("\n")
		return 1
	}
	// Synchronous call
	var args = querier.Args{Data: arg, Filepath: "/home/jbahary2/" + filepath}
	var reply string
	err = client.Call("Querier.Grep", args, &reply)
	if err != nil {
		return 2
	}
	fmt.Printf("%s\n", reply)
	return 0
}