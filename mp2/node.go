package main

import (
    "log"
    "net"
    "time"
    "fmt"
    "encoding/json"
    "detector"
    "hash/fnv"
    "memtable"
    "bytes"
    "mylog"
    // "beat_table"
    "beat_table"
    "sync"
)

type IntroMsg struct{
    node_hash int
    table memtable.Memtable
}

var isintroducer = false;

var node_id detector.Node_id_t
var node_hash = -1
// Add mem table declaration here
var message_hashes_mutex = &sync.Mutex{}
var message_hashes = make(map[int]int64)
var mem_table memtable.Memtable = memtable.NewMemtable()

//Beat Table
var beatable beat_table.Beat_table = beat_table.NewBeatTable()
var neigh [4]int = [4]int{-1,-1,-1,-1}

const portNum = "6000"
const introducer_hash = 0
const introducer_ip =  "192.168.81.13" // CHANGE LATER
const time_to_live = 4

const MESSAGE_EXPIRE_TIME_MILLIS = 6000 // in milliseconds
const REDUNDANCY_TABLE_CLEAR_TIME_MILLIS = 6000 // in milliseconds
const HEARTBEAT_INTERVAL_MILLIS = 1000 // in milliseconds
func sendmessage(msg_struct detector.Msg_t, ip_raw net.IP, portNum string) {
    msg, err := json.Marshal(msg_struct)
    if err != nil {
        log.Fatal(err)
    }
    ip := ip_raw.String()
    service := ip + ":" + portNum
    fmt.Println("(sendmessage) SERVICE: %s", service)
    remoteaddr , err := net.ResolveUDPAddr("udp", service)
    if err != nil {
        log.Fatal(err)
    }
    conn, err := net.DialUDP("udp", nil, remoteaddr)

    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()
    _ , err = conn.Write([]byte(msg))
    if err != nil {
        log.Fatal(err)
    }
}

func unmarshalmsg(buf []byte) detector.Msg_t{
    msg := detector.Msg_t{}
    err := json.Unmarshal(bytes.Trim(buf, "\x00"), &msg)
    if err != nil {
        log.Fatal(err)
        // return nil
    }
    return msg
}

func handlejoinreqmsg(msg detector.Msg_t, addr *net.UDPAddr) {
    if isintroducer {
        hash := mem_table.Get_avail_hash()
        // neighbors := mem_table.Get_neighbors(introducer_hash)

        // add the node to the introducers table
        mem_table.Add_node(hash, msg.Node_id)
        neigh = beatable.Reval_table(node_hash, mem_table)

        // send this new node its hash_id and membership list
        sendintroinfo(hash, mem_table, addr)

        // send a join message to 2 previous and next nodes
        for i := 0; i <= len(neigh); i++ {
            neighbor_id := mem_table.Get_node(neigh[i])
            // Node id is generated in the msg
            mesg := detector.Msg_t{detector.JOIN, time.Now().UnixNano(), msg.Node_id, time_to_live, byte(hash)}
            sendmessage(mesg, neighbor_id.IPV4_addr, portNum)
        }
    } else {
        log .Fatal("Only introducer should receive JOIN_REQUESTS. Ignoring")
    }
}

func sendintroinfo(node_hash int, mem_table memtable.Memtable, addr *net.UDPAddr){
    msg_struct := IntroMsg{node_hash, mem_table}
    msg, err := json.Marshal(msg_struct)
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    _ , err = conn.Write([]byte(msg))
    if err != nil {
        log.Fatal(err)
    }
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
        neigh = beatable.Reval_table(node_hash, mem_table)

        // add it to the map, and then process it
        addtomessagehashes(hash_msg)
        // dont continue to send if no more jumps left
        if msg.Time_to_live <= 0 {
            return
        }

        msg.Time_to_live -= 1

        neighbors := mem_table.Get_neighbors(node_hash)
        for i := 0; i <= len(neighbors); i++ {
            neighbor_id := mem_table.Get_node(neighbors[i])
            sendmessage(msg, neighbor_id.IPV4_addr, portNum)
        }
    }
}

//TODO
func handleheartbeatmsg(msg detector.Msg_t) {
}

func handlefailmsg(msg detector.Msg_t) {

}

func handleleavemsg(msg detector.Msg_t) {
    hash_msg := hashmsgstruct(msg)
    exists := findkeyinmessagehashes(hash_msg)
    if !exists {

        // delete the node from table
        mem_table.Delete_node(int(msg.Node_hash), msg.Node_id)
        neigh = beatable.Reval_table(node_hash, mem_table)

        addtomessagehashes(hash_msg)

        if msg.Time_to_live <= 0 {
            return
        }

        msg.Time_to_live -= 1

        neighbors := mem_table.Get_neighbors(node_hash)
        for i := 0; i <= len(neighbors); i++ {
            neighbor_id := mem_table.Get_node(neighbors[i])
            sendmessage(msg, neighbor_id.IPV4_addr, portNum)
        }
    }
}

func handleconnection(buffer []byte, addr *net.UDPAddr) {
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
    hostName := "localhost"
    service := hostName + ":" + portNum
    udpAddr, err := net.ResolveUDPAddr("udp4", service)
    if err != nil {
        log.Fatal(err)
    }
    // setup listener for incoming UDP connection
    ln, err := net.ListenUDP("udp", udpAddr)
    if err != nil {
        log.Fatal(err)
    }
    mylog.Log_writeln("UDP server up and listening on port " + portNum)
    defer ln.Close()
    for {
        buffer := make([]byte, 1024)
        // wait for UDP client to connect
        _, addr, err := ln.ReadFromUDP(buffer)

        if err != nil {
            log.Fatal(err)
            continue
        }
        mylog.Log_writeln("Found new connection")
        go handleconnection(buffer, addr)
    }
}

func init_() {
    mylog.Log_init()
    node_id = detector.Gen_node_id()
    if node_id.IPV4_addr.String() == introducer_ip {
        isintroducer = true
        node_hash = introducer_hash
    } else {
        intro_info := join_cluster(node_id)
        node_hash = intro_info.node_hash
        mem_table = intro_info.table
        neigh = beatable.Reval_table(node_hash, mem_table)   
    }
}

func join_cluster(node_id detector.Node_id_t) IntroMsg{
    //Start up a server to receive back response
    hostName := "localhost"
    service := hostName + ":" + portNum
    udpAddr, err := net.ResolveUDPAddr("udp4", service)
    if err != nil {
        log.Fatal(err)
    }
    // setup listener for incoming UDP connection
    ln, err := net.ListenUDP("udp", udpAddr)
    if err != nil {
        log.Fatal(err)
    }
    mylog.Log_writeln("UDP server up and listening on port " + portNum)
    buffer := make([]byte, 2048)
    defer ln.Close()
    for{
        fmt.Printf("Messaging Introducer . . .")
        //Message the introducer
        msg_struct := detector.Msg_t{detector.JOIN_REQ, time.Now().UnixNano(), node_id, byte(time_to_live), byte(node_hash)}
        sendmessageintroducer(msg_struct, portNum)
        //Wait for reply
        //Set a deadline (play around with time to not duplicate message)
        ln.SetReadDeadline(time.Now().Add(10000))
        // wait for UDP client to connect
        _, _, err = ln.ReadFromUDP(buffer)
        if err != nil {
            //We timed out might be the error
            fmt.Printf("Introducer never responded, trying again . . .")
            continue
        }
        mylog.Log_writeln("Introducer has responded!")
        msg := IntroMsg{}
        err = json.Unmarshal(bytes.Trim(buffer, "\x00"), &msg)
        if err != nil {
            log.Fatal(err)
        }
        return msg
    }
    return IntroMsg{}
}


func sendmessageintroducer(msg_struct detector.Msg_t, portNum string) {
    msg, err := json.Marshal(msg_struct)
    if err != nil {
        log.Fatal(err)
    }
    ip := introducer_ip
    service := ip + ":" + portNum
    fmt.Println("(sendmessage) SERVICE: %s", service)
    remoteaddr , err := net.ResolveUDPAddr("udp", service)
    if err != nil {
        log.Fatal(err)
    }
    conn, err := net.DialUDP("udp", nil, remoteaddr)

    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()
    _ , err = conn.Write([]byte(msg))
    if err != nil {
        log.Fatal(err)
    }
}

func heartbeatsend() {
        for {
            neigh = beatable.Reval_table(node_hash, mem_table)
            for i := 0; i <= len(neigh); i++ {
                neighbor_id := mem_table.Get_node(neigh[i])
                // Node id is generated in the msg
                mesg := detector.Msg_t{detector.HEARTBEAT, time.Now().UnixNano(), node_id, time_to_live, byte(node_hash)}
                sendmessage(mesg, neighbor_id.IPV4_addr, portNum)
            }
            time.Sleep(HEARTBEAT_INTERVAL_MILLIS * time.Millisecond)
        }
}

func main() {
    // init_()
    go listener()
    go heartbeatsend()
    for{
        message_hashes_mutex.Lock()
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


