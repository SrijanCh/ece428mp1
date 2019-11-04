package main
import(
	"sdfsrpc"
	"log"
	"net/rpc"
	// "net/http"
	"net"
	"fmt"
	// "time"
	"os"
    "strings"
    "bufio"
	// "zookeeper"
    "io/ioutil"
	// "log"	
)

var zoo_ip string = "172.22.156.255"//"172.22.154.255"
const zoo_portnum = "3075"
const node_portnum = "3074"

func parse_command(command string) {
    split_command := strings.Split(command, " ")
    // Trim the newline because go...
    for i := 0; i < len(split_command); i++ {
        split_command[i] = strings.TrimSpace(split_command[i])
    }
    if "put" == split_command[0] {
        if len(split_command) != 3 {
            fmt.Printf("Invalid number of args for put\n")
        } else {
            if put_confirm_req(split_command[2], zoo_ip, zoo_portnum) {
                fmt.Printf("Are you sure you want to replace recently put file %s? (y/n)\n", split_command[2])
                reader := bufio.NewReader(os.Stdin)
                text, _ := reader.ReadString('\n')
                text = strings.TrimSpace(text)
                if text == "y" {
                    fmt.Printf("Request confirmed\n")
                    c_put(split_command[1], split_command[2])
                } else if text == "n" {
                    fmt.Printf("Request cancelled\n")
                } else {
                    fmt.Printf("Invalid response...cancelling\n")
                }
            } else {
                // Otherwise just do the request. No need to warn user
                c_put(split_command[1], split_command[2])
            }
        }
    } else if "get" == split_command[0] {
        if len(split_command) != 3 {
            fmt.Printf("Invalid number of args for get\n")
        } else {
            c_get(split_command[1], split_command[2])
        }
    } else if "delete" == split_command[0] {

        if len(split_command) != 2 {
            fmt.Printf("Invalid number of args for delete\n")
        } else {
            c_delete(split_command[1])
        }
    } else if "ls" == split_command[0] {

        if len(split_command) != 2 {
            fmt.Printf("Invalid number of args for ls\n");
        } else {
            c_ls(split_command[1])
        }
    } else if "store" == split_command[0] {
        if len(split_command) != 2 {
            fmt.Printf("Invalid number of args for store\n")
        } else {
            c_store(split_command[1])
        }
    } else {
        fmt.Printf("Invalid command\n")
    }
}

func main(){
    reader := bufio.NewReader(os.Stdin)
    for {
        text, _ := reader.ReadString('\n')
        parse_command(text)
    }
	// write("testfile", "abcdefg", "172.22.154.255", "3074")
	// var a sdfsrpc.Read_reply = read("testfile", "172.22.154.255", "3074")
	// fmt.Printf("Read %s with Timestamp %d\n", a.Data, a.Timestamp)
	// write("testfile", "NEW DATA\n", "172.22.154.255", "3074")
	// a = read("testfile", "172.22.154.255", "3074")
	// fmt.Printf("Read %s with Timestamp %d\n", a.Data, a.Timestamp)
	// // delete("testfile", "172.22.154.255", "3074")
	// rep_to("testfile", "10.192.103.233", "3074", "172.22.154.255", "3074")
	c_put("log.txt", "firstfile")
}


func c_put(localname, sdfsname string){
	fmt.Printf("PUT %s %s\n", localname, sdfsname)

	fmt.Printf("Reading from %s...\n", localname)
	file, err := os.Open(localname)
	if(err != nil){
		fmt.Printf("Local file open error: %s\n", localname)
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
  		fmt.Printf("Error with file stats\n")// Could not obtain stat, handle error
	}
    
    ips, timestamp := put_req(sdfsname, getmyIP(), zoo_ip, zoo_portnum)

	if fi.Size() < 5605500{
		fmt.Printf("Sending file over in one go\n")
    	Data_r, _ := ioutil.ReadAll(file)
	
		fmt.Printf("Contacting zookeeper with %s to get IPs...\n", sdfsname)
    	// ips, timestamp := put_req(sdfsname, getmyIP(), zoo_ip, zoo_portnum)
    	
		fmt.Printf("Got IPs! Copying file to IPs...\n")
    	//Write the file to the four nodes
    	for _,node_ip := range ips{
    		fmt.Printf("Broadcasting to %s...\n", node_ip)
    		write(sdfsname, timestamp, string(Data_r), node_ip, node_portnum)
    	}
    	fmt.Printf("Done dispatching PUT broadcasts.\n")
	}else{
		fmt.Printf("Too huge! Sending file over in chunks\n")
    	var firstwrite = true
    	n := 1
    	buf := make([]byte, 15000)
    	reader := bufio.NewReader(file)
    	// _, err := ioutil.ReadAll(file)
    	for n != 0 {
    		n, _ = reader.Read(buf) 
			// fmt.Printf("Contacting zookeeper with %s to get IPs...\n", sdfsname)
    		// ips, timestamp := put_req(sdfsname, getmyIP(), zoo_ip, zoo_portnum)
    		
			// fmt.Printf("Got IPs! Copying file to IPs...\n")
    		//Write the file to the four nodes
    		for _,node_ip := range ips{
    			fmt.Printf("Broadcasting a chunk of size %d to %s...\n", n, node_ip)
    			if firstwrite{
    				fmt.Printf("Create the file\n")
    				write(sdfsname, timestamp, string(buf), node_ip, node_portnum)
    				firstwrite = false
    			}else{
    				fmt.Printf("Appending to the file\n")
    				append(sdfsname, timestamp, string(buf), node_ip, node_portnum)
    			}
    		}
    	}

    	fmt.Printf("Done dispatching PUT broadcasts.\n")

	}
}

func c_get(sdfsname, localname string){
	fmt.Printf("GET: %s, %s\n", sdfsname, localname)
    fmt.Printf("Obtaining best node to get the file from...")

    get_reply := get_req(sdfsname, zoo_ip, zoo_portnum)
    fmt.Printf("Got node@%s; Reading file %s from it...\n", get_reply, sdfsname)

    read_reply := read(sdfsname, get_reply, node_portnum)
    fmt.Printf("Got the data from sdfs %s from node@%s; copying into local %s.\n", sdfsname, get_reply, localname)

	file, err := os.Create(localname)
	if err != nil{
		fmt.Printf("File creation error: %s\n", localname)
		return
	}

	defer file.Close()

	fmt.Printf("Writing to %s...\n")
    _, err = file.WriteString(read_reply.Data)
	if(err != nil){
		fmt.Printf("File write error: %s\n", localname)
		return
	}

	fmt.Printf("Done! Read the file %s from SDFS into %s\n", sdfsname, localname)
}

func c_delete(sdfsname string){
	fmt.Printf("Requesting delete of %s\n", sdfsname)
	del_req(sdfsname, zoo_ip, zoo_portnum)
}

func c_ls(sdfsname string){
	fmt.Printf("Checking where %s is located...\n", sdfsname)
	fmt.Printf("%s\n", ls_req(sdfsname, zoo_ip, zoo_portnum))
}

func c_store(ip string){
    fmt.Printf("Getting all file names at node ip %s\n", ip)
    fmt.Printf("%s\n", store_req(ip, zoo_ip, zoo_portnum))
}

type Put_args struct{
	Sdfsname, Call_ip string
}

type Put_return struct{
	Ips []string
	Timestamp int64
}

func put_req(sdfsname, my_ip, ip, port string) ([]string, int64){
	fmt.Printf("------------Put req----------------")
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	// Synchronous call***************************************
	var args = Put_args{Sdfsname: sdfsname, Call_ip: my_ip} //Create the args gob to send over to the RPC
	var reply Put_return //Create a container for our results
	fmt.Printf("Put_req args: %s, %s", sdfsname, my_ip)
	err = client.Call("Zookeeper.Zoo_put", args, &reply) //Make the remote call
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("Put_req reply: \n")
	for j,_ := range reply.Ips{
		fmt.Printf("%s\n", (reply.Ips)[j])
	}

	return reply.Ips, reply.Timestamp
}


type Get_args struct{
	Sdfsname string
}

type Get_return struct{
	Ip string
}

func get_req(sdfsname, ip, port string) (string){
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	// Synchronous call***************************************
	var args = Get_args{Sdfsname: sdfsname} //Create the args gob to send over to the RPC
	var reply Get_return //Create a container for our results
	err = client.Call("Zookeeper.Zoo_get", args, &reply) //Make the remote call
	if err != nil{
		log.Fatal(err)
	}
	return reply.Ip
}

type Del_args struct{
	Sdfsname string
}

func del_req(sdfsname, ip, port string) int64{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	// Synchronous call***************************************
	var args = Del_args{Sdfsname: sdfsname} //Create the args gob to send over to the RPC
	var reply int64 //Create a container for our results
	err = client.Call("Zookeeper.Zoo_del", args, &reply) //Make the remote call
	if err != nil{
		log.Fatal(err)
	}
	return reply
}

type Ls_args struct{
	Sdfsname string
}

func ls_req(sdfsname, ip, port string) string{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	// Synchronous call***************************************
	var args = Ls_args{Sdfsname: sdfsname} //Create the args gob to send over to the RPC
	var reply string //Create a container for our results
	err = client.Call("Zookeeper.Zoo_ls", args, &reply) //Make the remote call
	if err != nil{
		log.Fatal(err)
	}
	return reply
}

type Store_args struct {
    Node_ip string
}

func store_req(node_ip, ip, port string) string {
    client, err := rpc.DialHTTP("tcp", ip + ":" + port)
    if err != nil {
        log.Fatal(err)
    }
    // no args needed
    var args = Store_args{Node_ip: node_ip}
    var reply string
    err = client.Call("Zookeeper.Zoo_store", args, &reply)
    if err != nil {
        log.Fatal(err)
    }
    return reply
}

type Put_confirm_args struct {
    Sdfsname string
}

func put_confirm_req(Filename, zoo_ip, port string) bool {
    client, err := rpc.DialHTTP("tcp", zoo_ip + ":" + port)
    if err != nil {
        log.Fatal(err)
    }
    var args = Put_confirm_args{Sdfsname: Filename}
    var reply bool
    err = client.Call("Zookeeper.Zoo_put_confirm", args, &reply)
    if err != nil {
        log.Fatal(err)
    }
    return reply
}

func write(filename string, ts int64, data, ip, port string) int{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	// Synchronous call***************************************
	var args = sdfsrpc.Write_args{Sdfsname: filename, Data: data, Timestamp: ts} //Create the args gob to send over to the RPC
	var reply int //Create a container for our results
	_ = client.Call("Sdfsrpc.Write_file", args, &reply) //Make the remote call
	return reply
}


func append(filename string, ts int64, data, ip, port string) int{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	defer client.Close()
	if err != nil {
		log.Fatal(err)
	}
	// Synchronous call***************************************
	var args = sdfsrpc.Write_args{Sdfsname: filename, Data: data, Timestamp: ts} //Create the args gob to send over to the RPC
	var reply int //Create a container for our results
	_ = client.Call("Sdfsrpc.Append_file", args, &reply) //Make the remote call
	return reply
}

func read(filename, ip, port string) sdfsrpc.Read_reply{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	var args = sdfsrpc.Read_args{Sdfsname: filename} //Create the args gob to send over to the RPC
	var reply sdfsrpc.Read_reply //Create a container for our results
	err = client.Call("Sdfsrpc.Get_file", args, &reply) //Make the remote call
	if err != nil{
		log.Fatal(err)
	}
	return reply
}

func delete(filename, ip, port string) int64{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}

	var args = sdfsrpc.Read_args{Sdfsname: filename} //Create the args gob to send over to the RPC
	var reply int64 //Create a container for our results
	_ = client.Call("Sdfsrpc.Delete_file", args, &reply) //Make the remote call
	return reply
}

func get_timestamp(filename, ip, port string) int64{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	var args = sdfsrpc.Read_args{Sdfsname: filename} //Create the args gob to send over to the RPC
	var reply int64 //Create a container for our results
	err = client.Call("Sdfsrpc.Get_timestamp", args, &reply) //Make the remote call	
	if err != nil{
		log.Fatal(err)
	}
	return reply
}

func get_store(ip, port string) string{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	var args int
	var store_list string
	err = client.Call("Sdfsrpc.Get_store", args, &store_list) //Make the remote call	
	if err != nil{
		log.Fatal(err)
	}
	return store_list
}

func rep_to(filename, ipto, portto, ip, port string) int{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	var args = sdfsrpc.Rep_args{filename, ipto, portto}
	var retval int
	err = client.Call("Sdfsrpc.Replicate_to", args, &retval) //Make the remote call	
	if err != nil{
		log.Fatal(err)
	}
	return retval
}

func getmyIP() (string) {
	var myIp string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalf("Cannot get my IP")
		os.Exit(1)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				myIp = ipnet.IP.String()
			}
		}
	}
	return myIp
}
