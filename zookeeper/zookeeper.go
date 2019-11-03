package zookeeper

import(
	"sdfsrpc"
  	"math/rand"
  	"time"
  	"fmt"
	"net/rpc"
)

const portnum = "3074"         

type MemberNode struct {
	Ip string
	Timestamp int64
	Alive bool
}

var MemberMap = make(map[int]*MemberNode)
var Garbage = make(map[int]bool)
// var FailQueue = make([]*MemberNode)

type Zookeeper int

type FileLoc struct{
	MemID 	  int
	Ip 		  string
	Timestamp int64
}

var FileTable = make(map[string]([4]FileLoc))

// Thus, the allowed file ops include:

// 1) put localfilename sdfsfilename (from local dir)
//		4 initial writes
//		3 updates (quorum)
//		Check if file was written to in the last 1 minute, and ask for confirm if it was
//		Wait 30 sec for confirm, else yeet
//		~~~~~~~~~~~~~~~~~~~~~~~~~~FOR ZOOKEEPER~~~~~~~~~~~~~~~~~~~~~~~~~~~
//		If initial write: 		pick 4 nodes and return IP list
//		If late update (>1):	pick 3 of 4 nodes and return IP list
// 		If early update (<1):	Ask for confirm; wait 30 secs, if confirm arrives, send 3 of 4 IPs, else send 0 IPs

// 2) get sdfsfilename localfilename (fetches to local dir)
//		2 fetches (quorum)
//		update lesser timestamp if timestamps not equal (write)


// 3) delete sdfsfilename
// 4) ls sdfsfilename: list all machine (VM) addresses where this file is currently
// being stored; 
// 5) store: At any machine, list all files currently being stored at this
// machine. Here sdfsfilename is an arbitrary string while localfilename is
// the Unix-style local file system name.

// Put is used to both insert the original file and
// update it (updates are comprehensive, i.e., they send the entire file, not just a
// block of it).
func num_live() int{
	fmt.Printf("num_live\n")
	printMemberMap();
	count := 0
	for _, v := range MemberMap{
		if v.Alive{	
			count++
		}
	}
	return count
}

func printMemberMap() {
	fmt.Printf("Zookeeper member list: [\n")
	for id, ele := range(MemberMap) {
		fmt.Printf("%d: %s, %d, %t\n", id, ele.Ip, ele.Timestamp, ele.Alive)
	}
	fmt.Printf("]\n")
}

func pick4() ([]int, []string){
	var ret_str []string = make([]string, 0)
	var ret_int []int = make([]int, 0)
	var a,b,c,d int

	if (num_live() < 4){
		fmt.Printf("Not enough live nodes to pick random, %d\n", num_live())
		for k, v := range MemberMap{
			if v.Alive{	
				ret_str = append(ret_str, v.Ip)
				ret_int = append(ret_int, k)
				fmt.Printf("Returning %s and %d in set\n", v.Ip, k)
			}
		}
		return ret_int, ret_str
	}

	fmt.Printf("Picking 4 random nodes\n")

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	a = r.Intn(len(MemberMap))
	MemNode, ok := MemberMap[a]
	for !ok || !(*MemNode).Alive{
		a = r.Intn(len(MemberMap))
	}
	ret_str = append(ret_str, (*MemNode).Ip)
	ret_int = append(ret_int, a)
	
	b = r.Intn(len(MemberMap))
	MemNode, ok = MemberMap[b]
	for !ok || !(*MemNode).Alive || b == a{
		b = r.Intn(len(MemberMap))
	}
	ret_str = append(ret_str, (*MemNode).Ip)
	ret_int = append(ret_int, b)


	c = r.Intn(len(MemberMap))
	MemNode, ok = MemberMap[c]
	for !ok || !(*MemNode).Alive || c == a || c == b{
		c = r.Intn(len(MemberMap))
	}
	ret_str = append(ret_str, (*MemNode).Ip)
	ret_int = append(ret_int, c)


	d = r.Intn(len(MemberMap))
	MemNode, ok = MemberMap[d]
	for !ok || !MemNode.Alive || d == a || d == b || d == c{
		d = r.Intn(len(MemberMap))
	}
	ret_str = append(ret_str, (*MemNode).Ip)
	ret_int = append(ret_int, d)

	fmt.Printf("Picked [(%d,%d) %s,%s] [(%d,%d) %s,%s] [(%d,%d) %s,%s] and [(%d,%d) %s,%s]\n", a, ret_int[0], 
																							   MemberMap[a].Ip, ret_str[0],
																							   b, ret_int[1],
																							   MemberMap[b].Ip, ret_str[1],
																							   c, ret_int[2],
																							   MemberMap[c].Ip, ret_str[2],
																							   d, ret_int[3],
																							   MemberMap[d].Ip, ret_str[3])
	return ret_int, ret_str
}

func pick3(fileloc_arr [4]FileLoc) ([]int, []string){
	var ret_str []string = make([]string, 3)
	var ret_int []int = make([]int, 3)
	//Pick one to not
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	a := r.Intn(len(fileloc_arr))
	for !MemberMap[fileloc_arr[a].MemID].Alive{
		a = r.Intn(len(fileloc_arr))
	}
	ret_int = append(ret_int, a)
	ret_str = append(ret_str, fileloc_arr[a].Ip)


	b := r.Intn(len(fileloc_arr))
	for !MemberMap[fileloc_arr[b].MemID].Alive || a==b{
		b = r.Intn(len(fileloc_arr))
	}
	ret_int = append(ret_int, b)
	ret_str = append(ret_str, fileloc_arr[b].Ip)


	c := r.Intn(len(fileloc_arr))
	for !MemberMap[fileloc_arr[c].MemID].Alive || c==a || c==b{
		c = r.Intn(len(fileloc_arr))
	}
	ret_int = append(ret_int, c)
	ret_str = append(ret_str, fileloc_arr[c].Ip)


	fmt.Printf("Picked [(%d, MemID %d) %s,%s] [(%d, MemID %d) %s,%s] [(%d, MemID %d) %s,%s]\n",  			
																			a, fileloc_arr[a].MemID,
																		    fileloc_arr[a].Ip, ret_str[0],
																		    b, fileloc_arr[b].MemID,
																		    fileloc_arr[b].Ip, ret_str[1],
																			c, fileloc_arr[c].MemID,
																			fileloc_arr[c].Ip, ret_str[2])
	return ret_int, ret_str
}

func pick2(fileloc_arr [4]FileLoc) ([]int, []string){
	var ret_str []string = make([]string, 2)
	var ret_int []int = make([]int, 2)
	//Pick one to not
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	a := r.Intn(len(fileloc_arr))
	for !MemberMap[fileloc_arr[a].MemID].Alive{
		a = r.Intn(len(fileloc_arr))
	}
	ret_int = append(ret_int, a)
	ret_str = append(ret_str, fileloc_arr[a].Ip)


	b := r.Intn(len(fileloc_arr))
	for !MemberMap[fileloc_arr[b].MemID].Alive || a==b{
		b = r.Intn(len(fileloc_arr))
	}
	ret_int = append(ret_int, b)
	ret_str = append(ret_str, fileloc_arr[b].Ip)

	
	fmt.Printf("Picked [(%d, MemID %d) %s,%s] [(%d, MemID %d) %s,%s]\n", 	a, fileloc_arr[a].MemID,
																		    fileloc_arr[a].Ip, ret_str[0],
																		    b, fileloc_arr[b].MemID,
																		    fileloc_arr[b].Ip, ret_str[1])
	return ret_int, ret_str
}

type Put_args struct{
	Sdfsname, Call_ip string
}

type Put_return struct{
	Ips []string
	Timestamp int64
}

// 1) put localfilename sdfsfilename (from local dir)
//		4 initial writes
//		3 updates (quorum)
//		Check if file was written to in the last 1 minute, and ask for confirm if it was
//		Wait 30 sec for confirm, else yeet
//		~~~~~~~~~~~~~~~~~~~~~~~~~~FOR ZOOKEEPER~~~~~~~~~~~~~~~~~~~~~~~~~~~
//		If initial write: 		pick 4 nodes and return IP list
//		If late update (>1):	pick 3 of 4 nodes and return IP list
// 		If early update (<1):	Ask for confirm; wait 30 secs, if confirm arrives, send 3 of 4 IPs, else send 0 IPs
func (t *Zookeeper) Zoo_put(args Put_args, reply *Put_return) error {

	//Just a check-in for a handled request
	if fileloc_arr, ok := FileTable[args.Sdfsname]; !ok { //NEW PUT
		a, b := pick4()
		c := int64(time.Now().Nanosecond())
		var f [4]FileLoc
		f[0] = FileLoc{a[0], b[0], c}
		f[1] = FileLoc{a[1], b[1], c}
		f[2] = FileLoc{a[2], b[2], c}
		f[3] = FileLoc{a[3], b[3], c}
		FileTable[args.Sdfsname] = f
		(*reply).Ips = b
		(*reply).Timestamp = c
	}else{									   //UPDATE (QUORUM)
		//Check on timestamp to see when last write was

		//Pick random 3 because Quorum
		a, b := pick3(fileloc_arr)
		c := int64(time.Now().Nanosecond())
		fileloc_arr[a[0]].Timestamp = c
		fileloc_arr[a[1]].Timestamp = c
		fileloc_arr[a[2]].Timestamp = c
		(*reply).Ips = b
		(*reply).Timestamp = c
	}

	// return nil
	//Form our return string, and then return
    // *reply = name + ":\n" + out.String() + "[" + str1 + "]\n";
	return nil
}

type Get_args struct{
	Sdfsname string
}

type Get_return struct{
	Ip string
}
// 2) get sdfsfilename localfilename (fetches to local dir)
//		2 fetches (quorum)
//		update lesser timestamp if timestamps not equal (write)
func (t *Zookeeper) Zoo_get(args Get_args, reply *Get_return) error {

	if fileloc_arr, ok := FileTable[args.Sdfsname]; !ok { //Invalid file
		fmt.Printf("GET: No such file found\n")
		(*reply).Ip = ""
	}else{
		_, b := pick2(fileloc_arr)
		c0 := get_timestamp(args.Sdfsname, b[0], portnum)
		c1 := get_timestamp(args.Sdfsname, b[1], portnum)
		if(c0 == c1){ 		//both are on same consistency
			(*reply).Ip = b[0]
		} else {				//One needs an update
			if c0 > c1 {
				(*reply).Ip = b[0]
				rep_to(args.Sdfsname, b[1], portnum, b[0], portnum)
			}else{
				(*reply).Ip = b[1]
				rep_to(args.Sdfsname, b[0], portnum, b[1], portnum)
			}
		}
	}

	return nil
}


type Del_args struct{
	Sdfsname string
}

// 3) delete sdfsfilename
func (t *Zookeeper) Zoo_del(args Del_args, reply *int64) error {
	if _, ok := FileTable[args.Sdfsname]; !ok { //Invalid file
		fmt.Printf("DEL: No such file found, so success I guess?\n")
	}else{
		for i := 0; i < 4; i++{
			del(args.Sdfsname, (FileTable[args.Sdfsname])[i].Ip, portnum)
		}
		delete(FileTable, args.Sdfsname)
	}
	(*reply) = 0
	return nil
}


// 4) ls sdfsfilename: list all machine (VM) addresses where this file is currently
// being stored;
type Ls_args struct{
	Sdfsname string
}

func (t* Zookeeper) Zoo_ls(args Ls_args, reply *string) error{
	str := ""
	if _, ok := FileTable[args.Sdfsname]; ok { //Invalid file
		fmt.Printf("LS: Starting search...\n")
		for i := 0; i < 4; i++{
			str += "(" + string((FileTable[args.Sdfsname])[i].MemID) + ")" + string((FileTable[args.Sdfsname])[i].Ip) + "\n"
		}
	}
	*reply = str
	return nil

}
// 5) store: At any machine, list all files currently being stored at this
// machine. Here sdfsfilename is an arbitrary string while localfilename is
// the Unix-style local file system name.

func del(filename, ip, port string) int64{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		return -1
	}

	var args = sdfsrpc.Read_args{Sdfsname: filename} //Create the args gob to send over to the RPC
	var reply int64 //Create a container for our results
	_ = client.Call("Sdfsrpc.Delete_file", args, &reply) //Make the remote call
	return reply
}

func get_timestamp(filename, ip, port string) int64{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		return -1
	}
	var args = sdfsrpc.Read_args{Sdfsname: filename} //Create the args gob to send over to the RPC
	var reply int64 //Create a container for our results
	_ = client.Call("Sdfsrpc.Get_timestamp", args, &reply) //Make the remote call	
	return reply
}

func rep_to(filename, ipto, portto, ip, port string) int{
	client, err := rpc.DialHTTP("tcp", ip + ":" + port) //Connect to given address
	if err != nil {
		return -1
	}
	var args = sdfsrpc.Rep_args{filename, ipto, portto}
	var retval int
	_ = client.Call("Sdfsrpc.Replicate_to", args, &retval) //Make the remote call	
	return retval
}
