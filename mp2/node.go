package main

import (
    //"log"
    "net"
    "time"
    "fmt"
    "encoding/json"
    "detector"
    "hash/fnv"
    "memtable"
    "bytes"
    "mylog"
    "beat_table"
    "sync"
)

type IntroMsg struct{
    node_hash int
    table memtable.Memtable
}

var isintroducer = false;

var my_node_id detector.Node_id_t
var my_node_hash = -1
// Add mem table declaration here
var message_hashes_mutex = &sync.Mutex{}
var message_hashes = make(map[int]int64)
var mem_table memtable.Memtable = memtable.NewMemtable()

//Beat Table
var beatable beat_table.Beat_table = beat_table.NewBeatTable()
var neigh [4]int = [4]int{-1,-1,-1,-1}

const portNum = "6000"
const portNumber = 6000
const introducer_hash = 0
const introducer_ip =  "172.22.154.255" // CHANGE LATER
const time_to_live = 4

const MESSAGE_EXPIRE_TIME_MILLIS = 6000 // in milliseconds
const REDUNDANCY_TABLE_CLEAR_TIME_MILLIS = 6000 // in milliseconds
const HEARTBEAT_INTERVAL_MILLIS = 1000 // in milliseconds
const MONITOR_PERIOD_MILLIS = 3000
func sendmessage(msg_struct detector.Msg_t, ip_raw net.IP, portNum string) {
    msg, err := json.Marshal(msg_struct)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        // log.Fatal(err)
    }
    ip := ip_raw.String()
    service := ip + ":" + portNum
    fmt.Println("(sendmessage) SERVICE: %s", service)
    remoteaddr , err := net.ResolveUDPAddr("udp", service)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        return
        // log.Fatal(err)
    }
    conn, err := net.DialUDP("udp", nil, remoteaddr)

    if err != nil {
        fmt.Printf("%s\n", err.Error())
        return
        // log.Fatal(err)
    }

    defer conn.Close()
    _ , err = conn.Write([]byte(msg))
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        // log.Fatal(err)
    }
}

func unmarshalmsg(buf []byte) detector.Msg_t{
    msg := detector.Msg_t{}
    err := json.Unmarshal(bytes.Trim(buf, "\x00"), &msg)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        // log.Fatal(err)
        // return nil
    }
    return msg
}

func handlejoinreqmsg(msg detector.Msg_t, addr *net.UDPAddr) {
    if isintroducer {
        hash := mem_table.Get_avail_hash()
        // neigh := mem_table.Get_neigh(introducer_hash)

        // add the node to the introducers table
        mem_table.Add_node(hash, msg.Node_id)
        neigh = beatable.Reval_table(my_node_hash, mem_table)
        
        // send this new node its hash_id and membership list
        sendintroinfo(hash, mem_table, addr)

        if neigh[0] == -1 || neigh[1] == -1 || neigh[2] == -1 || neigh[3] == -1{
            // fmt.Printf("Can't get neigh\n")
            return
        }
        // send a join message to 2 previous and next nodes
        for i := 0; i <= len(neigh); i++ {
            neighbor_id := mem_table.Get_node(neigh[i])
            // Node id is generated in the msg
            mesg := detector.Msg_t{detector.JOIN, time.Now().UnixNano(), msg.Node_id, time_to_live, byte(hash)}
            sendmessage(mesg, neighbor_id.IPV4_addr, portNum)
        }
    } else {
        // log .Fatal("Only introducer should receive JOIN_REQUESTS. Ignoring")
    }
}


func sendmessageintroducer(msg_struct detector.Msg_t, portNum string) {
    msg, err := json.Marshal(msg_struct)
    if err != nil {
        mylog.Log_writeln("[sendmessageintroducer] Failed to marshal")
        fmt.Printf("%s\n", err.Error())
        // log.Fatal(err)
    }
    ip := introducer_ip
    service := ip + ":" + portNum
    fmt.Println("(sendmessage) SERVICE: %s", service)
    remoteaddr , err := net.ResolveUDPAddr("udp", service)
    if err != nil {
        mylog.Log_writeln("[sendmessageintroducer] Failed to get remote address")
        fmt.Printf("%s\n", err.Error())
        return
        // log.Fatal(err)
    }
    conn, err := net.DialUDP("udp", nil, remoteaddr)

    if err != nil {
        mylog.Log_writeln("[sendmessageintroducer] Failed to dial address")
        fmt.Printf("%s\n", err.Error())
        return
        // log.Fatal(err)
    }

    defer conn.Close()
    _ , err = conn.Write([]byte(msg))
    if err != nil {
        mylog.Log_writeln("[sendmessageintroducer] Failed to send message")
        fmt.Printf("%s\n", err.Error())
        // log.Fatal(err)
    }
}

func sendintroinfo(node_hash int, pass_mem_table memtable.Memtable, addr *net.UDPAddr){

    (*addr).Port = portNumber
    fmt.Printf("Got to sendintroinfo, sending hash %d and memtable below to %s\n", (*addr).String())
    fmt.Printf(mem_table.String())
    msg_struct := IntroMsg{node_hash, pass_mem_table}
    msg, err := json.Marshal(msg_struct)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        // log.Fatal(err)
    }
    fmt.Printf("Post-Marshalled msg: %s\n", string(msg))
    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        return
        // log.Fatal(err)
    }

    defer conn.Close()

    _ , err = conn.Write([]byte(msg))
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        // log.Fatal(err)
    }
    fmt.Printf("Ended sendintroinfo\n")
}

// a function that returns the hash of a structs members except the time to live
func hashmsgstruct(msg detector.Msg_t) int{
    s := string(msg.Msg_type) + string(msg.Timestamp) + string(msg.Node_id.Timestamp) + string(msg.Node_id.IPV4_addr) + string(msg.Node_hash)
    h := fnv.New32a()
    h.Write([]byte(s))
    return int(h.Sum32())
}

func addtomessagehashes(hash int) {
    message_hashes_mutex.Lock()
    message_hashes[hash] = time.Now().UnixNano()
    message_hashes_mutex.Unlock()
}

func findkeyinmessagehashes(hash int) bool {
    message_hashes_mutex.Lock()
    _, exists := message_hashes[hash]
    message_hashes_mutex.Unlock()
    return exists
}

func handlejoinmsg(msg detector.Msg_t) {
    hash_msg := hashmsgstruct(msg)
    exists := findkeyinmessagehashes(hash_msg)
    if !exists {

        // add the node to the table
        mem_table.Add_node(int(msg.Node_hash), msg.Node_id)
        neigh = beatable.Reval_table(my_node_hash, mem_table)
        
        if neigh[0] == -1 || neigh[1] == -1 || neigh[2] == -1 || neigh[3] == -1{
            fmt.Printf("Can't get neigh\n")
            return
        }


        // add it to the map, and then process it
        addtomessagehashes(hash_msg)
        // dont continue to send if no more jumps left
        if msg.Time_to_live <= 0 {
            return
        }

        msg.Time_to_live -= 1

        // neigh := mem_table.Get_neigh(my_node_hash)
        for i := 0; i <= len(neigh); i++ {
            neighbor_id := mem_table.Get_node(neigh[i])
            sendmessage(msg, neighbor_id.IPV4_addr, portNum)
        }
    }
}

//TODO
func handleheartbeatmsg(msg detector.Msg_t) {
    beatable.Log_beat(int(msg.Node_hash), msg.Timestamp)
}

func handlefailmsg(msg detector.Msg_t) {
    mylog.Log_writeln("[handlefailmsg] Relaying failure message")
    hash_msg := hashmsgstruct(msg)
    exists := findkeyinmessagehashes(hash_msg)
    if !exists {
        
        mem_table.Delete_node(int(msg.Node_hash), msg.Node_id)
        

        neigh = beatable.Reval_table(my_node_hash, mem_table)
        
        addtomessagehashes(hash_msg)

        if msg.Time_to_live <= 0 {
            return
        }

        msg.Time_to_live -= 1

        for i := 0; i < len(neigh); i++ {
            neighbor_id := mem_table.Get_node(neigh[i])
            sendmessage(msg, neighbor_id.IPV4_addr, portNum)
        }
    }
}

func handleleavemsg(msg detector.Msg_t) {
    mylog.Log_writeln("[handleleavemsg] Leaving the network")
    hash_msg := hashmsgstruct(msg)
    exists := findkeyinmessagehashes(hash_msg)
    if !exists {

        // delete the node from table
        mem_table.Delete_node(int(msg.Node_hash), msg.Node_id)
        neigh = beatable.Reval_table(my_node_hash, mem_table)
        
        if neigh[0] == -1 || neigh[1] == -1 || neigh[2] == -1 || neigh[3] == -1{
            fmt.Printf("Can't get neigh\n")
            return
        }


        addtomessagehashes(hash_msg)

        if msg.Time_to_live <= 0 {
            return
        }

        msg.Time_to_live -= 1

        // neigh := mem_table.Get_neigh(my_node_hash)
        for i := 0; i <= len(neigh); i++ {
            neighbor_id := mem_table.Get_node(neigh[i])
            sendmessage(msg, neighbor_id.IPV4_addr, portNum)
        }
    }
}

func handleconnection(buffer []byte, addr *net.UDPAddr) {
    mylog.Log_writeln("[handleconnection] Got new connection")
    msg := unmarshalmsg(buffer)

    switch msg.Msg_type {
        case detector.HEARTBEAT:
            handleheartbeatmsg(msg)
        case detector.JOIN:
            handlejoinmsg(msg)
        case detector.FAIL:
            handlefailmsg(msg)
        case detector.LEAVE:
            handleleavemsg(msg)
        case detector.JOIN_REQ:
            handlejoinreqmsg(msg, addr)
        default:
            // handlemisc(msg)
    }
}

func listener() {
    // hostName := "localhost"
    hostName := my_node_id.IPV4_addr.String()
    service := hostName + ":" + portNum
    udpAddr, err := net.ResolveUDPAddr("udp4", service)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        return
        // log.Fatal(err)
    }
    // setup listener for incoming UDP connection
    ln, err := net.ListenUDP("udp", udpAddr)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        return
        // log.Fatal(err)
    }
    fmt.Printf("UDP server up and listening on addr " + hostName + ", port " + portNum + "\n")
    mylog.Log_writeln("UDP server up and listening on port " + portNum)
    defer ln.Close()
    for {
        buffer := make([]byte, 1024)
        // wait for UDP client to connect
        _, addr, err := ln.ReadFromUDP(buffer)

        if err != nil {
            fmt.Printf("%s\n", err.Error())
            // log.Fatal(err)
            continue
        }
        mylog.Log_writeln("Found new connection")
        go handleconnection(buffer, addr)
    }
}

func monitor(){
    mylog.Log_writeln("[monitor] Starting up monitor...")
    var stamps = [4]int64{-1,-1,-1,-1}
    var fails []int
    for{
        time.Sleep(MONITOR_PERIOD_MILLIS * time.Millisecond)

        if neigh[0] == -1 || neigh[1] == -1 || neigh[2] == -1 || neigh[3] == -1{
          // fmt.Printf("Can't get neigh\n")
            continue
        }

        for i:=0; i < len(neigh); i++{
            a := beatable.Get_beat(neigh[i])
            if stamps[i] == a{
                fails = append(fails, neigh[i])
            }else{
                stamps[i] = a
            }
        }

        for i:=0; i < len(fails); i++{
            declare_fail(fails[i])
        }
    }
}

func declare_fail(node_hash int){
    mylog.Log_writeln("[declare_fail] Sending out failure messages...")

    // delete the node from table
    a := mem_table.Get_node(node_hash)
    mem_table.Delete_node(int(node_hash), mem_table.Get_node(node_hash))
    neigh = beatable.Reval_table(my_node_hash, mem_table)
        
        if neigh[0] == -1 || neigh[1] == -1 || neigh[2] == -1 || neigh[3] == -1{
            // fmt.Printf("Can't get neigh\n")
            return
        }


    msg := detector.Msg_t{detector.FAIL, time.Now().UnixNano(), a, byte(time_to_live), byte(node_hash)}

    // neigh := mem_table.Get_neigh(my_node_hash)
    for i := 0; i <= len(neigh); i++ {
            neighbor_id := mem_table.Get_node(neigh[i])
            sendmessage(msg, neighbor_id.IPV4_addr, portNum)
    }
}

func init_() {
    mylog.Log_init()
    my_node_id = detector.Gen_node_id()
    if my_node_id.IPV4_addr.String() == introducer_ip {
        fmt.Print("We are the introducer\n")
        isintroducer = true
        my_node_hash = introducer_hash
        mem_table.Add_node(my_node_hash, my_node_id)
        neigh = beatable.Reval_table(my_node_hash, mem_table)  
    } else {
        intro_info := join_cluster(my_node_id)
        my_node_hash = intro_info.node_hash
        mem_table = intro_info.table
        neigh = beatable.Reval_table(my_node_hash, mem_table)   
    }       
    fmt.Printf("Our node is initialized! This node is hashed to %d with node_id %s:%d.\n", my_node_hash, my_node_id.IPV4_addr.String(), my_node_id.Timestamp)
    fmt.Printf("Our membership table currently looks as such:\n %s.\n", mem_table.String())

    if neigh[0] == -1 || neigh[1] == -1 || neigh[2] == -1 || neigh[3] == -1{
        fmt.Printf("Can't get neigh\n")
        return
    }

}

func join_cluster(node_id detector.Node_id_t) IntroMsg{
    //Start up a server to receive back response
    a := 0
    hostName := my_node_id.IPV4_addr.String()
    service := hostName + ":" + portNum
    fmt.Printf("Introducer listening server, at service %s\n", service)
    udpAddr, err := net.ResolveUDPAddr("udp4", service)
    if err != nil {
        mylog.Log_writeln("[join_cluster] Failed to resolve UDP address")
        fmt.Printf("%s\n", err.Error())
        return IntroMsg{}
        // log.Fatal(err)
    }
    // setup listener for incoming UDP connection
    ln, err := net.ListenUDP("udp", udpAddr)
    if err != nil {
        mylog.Log_writeln("[join_cluster] Failed to get listener")
        fmt.Printf("%s\n", err.Error())
        return IntroMsg{}
        // log.Fatal(err)
    }
    mylog.Log_writeln("UDP server up and listening on port " + portNum)
    buffer := make([]byte, 2048)
    defer ln.Close()
    for{
        mylog.Log_writeln("[join_cluster] Messaging introducer...")
        fmt.Printf("Messaging Introducer . . .\n")
        //Message the introducer
        msg_struct := detector.Msg_t{detector.JOIN_REQ, time.Now().UnixNano(), node_id, byte(time_to_live), byte(my_node_hash)}
        sendmessageintroducer(msg_struct, portNum)
        //Wait for reply
        //Set a deadline (play around with time to not duplicate message)
        ln.SetReadDeadline(time.Now().Add(time.Millisecond * 10000))
        // wait for UDP client to connect
        _, _, err = ln.ReadFromUDP(buffer)
        if err != nil {
            //We timed out might be the error
            mylog.Log_writeln("[join_cluster] Introducer never connected/responded, trying again...")
            fmt.Printf("(%d) Introducer never connected/responded, trying again . . .\n", a)
            a++
            continue
        }
        mylog.Log_writeln("Introducer has responded!")
        fmt.Printf("Pre-unmarshalled msg: %s\n", string(bytes.Trim(buffer, "\x00")))
        msg := IntroMsg{}
        err = json.Unmarshal(bytes.Trim(buffer, "\x00"), &msg)
        if err != nil {
            mylog.Log_writeln("[join_cluster] Failed to unmarshal")
            fmt.Printf("%s\n", err.Error())
            // log.Fatal(err)
        }
        return msg
    }
    return IntroMsg{}
}

func heartbeatsend() {
        for {
            neigh = beatable.Reval_table(my_node_hash, mem_table)

            if neigh[0] == -1 || neigh[1] == -1 || neigh[2] == -1 || neigh[3] == -1{
                // fmt.Printf("heartbeatsend: Can't get neigh\n")
                continue
            }

            mylog.Log_writeln("Sending heartbeat") 
            for i := 0; i <= len(neigh); i++ {
                neighbor_id := mem_table.Get_node(neigh[i])
                // Node id is generated in the msg
                mesg := detector.Msg_t{detector.HEARTBEAT, time.Now().UnixNano(), my_node_id, time_to_live, byte(my_node_hash)}
                sendmessage(mesg, neighbor_id.IPV4_addr, portNum)
            }
            time.Sleep(HEARTBEAT_INTERVAL_MILLIS * time.Millisecond)
        }
}

func main() {
    mylog.Log_init()
    init_()
    go listener()
    go heartbeatsend()
    go monitor()
    for{
        message_hashes_mutex.Lock()
        mylog.Log_writeln("[main] Flushing redundancy map")
        for k, e := range message_hashes {
            t := time.Now().UnixNano()
            // measured in millis
            if ((t - e) / 1000000) >= MESSAGE_EXPIRE_TIME_MILLIS {
                delete(message_hashes, k)
            }
        }
        message_hashes_mutex.Unlock()
        time.Sleep(REDUNDANCY_TABLE_CLEAR_TIME_MILLIS * time.Millisecond)
    }
}


