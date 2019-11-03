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
	// "zookeeper"
    "io/ioutil"
	// "log"	
)

var zoo_ip string = "172.22.154.255"
const zoo_portnum = "3075"
const node_portnum = "3074"

func main(){
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
    Data_r, err := ioutil.ReadAll(file)

	fmt.Printf("Contacting zookeeper with %s to get IPs...\n", sdfsname)
    ips, timestamp := put_req(sdfsname, getmyIP(), zoo_ip, zoo_portnum)
    
	fmt.Printf("Got IPs! Copying file to IPs...\n")
    //Write the file to the four nodes
    for _,node_ip := range ips{
    	fmt.Printf("Broadcasting to %s...\n", node_ip)
    	write(sdfsname, timestamp, string(Data_r), node_ip, node_portnum)
    }
    fmt.Printf("Done dispatching PUT broadcasts.\n")
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