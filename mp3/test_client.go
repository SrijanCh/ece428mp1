package main
import(
	"sdfsrpc"
	"log"
	"net/rpc"
	// "net/http"
	// "net"
	"fmt"
	"time"
)

func main(){
	// client, err := rpc.DialHTTP("tcp", "127.0.0.1" + ":3074") //Connect to given address
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// Synchronous call
	// var args = sdfsrpc.Write_args{Sdfsname: "testfile", Data: "Abcd\n", Timestamp: 1234} //Create the args gob to send over to the RPC
	// var reply string //Create a container for our results
	// _ = client.Call("Sdfsrpc.Write_file", args, &reply) //Make the remote call

	// var args2 = sdfsrpc.Read_args{Sdfsname: "testfile"} //Create the args gob to send over to the RPC
	// var reply2 sdfsrpc.Read_reply //Create a container for our results
	// _ = client.Call("Sdfsrpc.Get_file", args2, &reply2) //Make the remote call
	// fmt.Printf("Data was %s with Timestamp %d\n", reply2.Data, reply2.Timestamp)


	// _ = client.Call("Sdfsrpc.Delete_file", args2, &reply) //Make the remote call
	// _ = client.Call("Sdfsrpc.Get_file", args2, &reply2) //Make the remote call
	// fmt.Printf("Data was %s with Timestamp %d after ", reply2.Data, reply2.Timestamp)
	
	//If the remote call fails report error
	// if err != nil {
	// 	return 2
	// }

	write("testfile", "abcdefg", "172.22.154.255", "3074")
	var a sdfsrpc.Read_reply = read("testfile", "172.22.154.255", "3074")
	fmt.Printf("Read %s with Timestamp %d\n", a.Data, a.Timestamp)
	write("testfile", "NEW DATA\n", "172.22.154.255", "3074")
	a = read("testfile", "172.22.154.255", "3074")
	fmt.Printf("Read %s with Timestamp %d\n", a.Data, a.Timestamp)
	// delete("testfile", "172.22.154.255", "3074")
	rep_to("testfile", "192.168.56.1", "3074", "172.22.154.255", "3074")
}

func write(filename, data, ip, port string) int{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	// Synchronous call***************************************
	var args = sdfsrpc.Write_args{Sdfsname: filename, Data: data, Timestamp: int64(time.Now().Nanosecond())} //Create the args gob to send over to the RPC
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
	_ = client.Call("Sdfsrpc.Get_file", args, &reply) //Make the remote call
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
	_ = client.Call("Sdfsrpc.Get_timestamp", args, &reply) //Make the remote call	
	return reply
}

func get_store(ip, port string) string{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	var args int
	var store_list string
	_ = client.Call("Sdfsrpc.Get_store", args, &store_list) //Make the remote call	
	return store_list
}

func rep_to(filename, ipto, portto, ip, port string) int{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		log.Fatal(err)
	}
	var args = sdfsrpc.Rep_args{filename, ipto, portto}
	var retval int
	_ = client.Call("Sdfsrpc.Replicate_to", args, &retval) //Make the remote call	
	return retval
}