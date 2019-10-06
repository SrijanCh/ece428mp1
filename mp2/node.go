package node

import (
    "log"
    "net"
    "time"
    "fmt"
    "encoding/json"
    "detector"
    "hash/fnv"
    "memtable"
    "beat_table"
)


var node_id = ""
var node_hash = -1
// Add mem table declaration here
var message_hashes = make(map[int]int)
var mem_table memtable.Memtable = NewMemtable()

const portNum = "6000"
const introducer_hash = 0
const introducer_ip =  "192.168.81.13" // CHANGE LATER
const time_to_live = 4

func sendmessage(msg_struct Msg_t, ip string, portNum string) {
    msg, err := json.Marshal(msg_struct)
    if err != nil {
        log.Fatal(err)
    }
    service := ip + ":" + portNum
    remoteaddr , err := net.ResolveUDPAddr("udp", service)
    if err != nil {
        log.Fatal(err)
    }
    conn, err := net.DialUDP("udp", nil, service)

    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()
    _ , err = conn.Write([]byte(msg))
    if err != nil {
        log.Fatal(err)
    }
}

func unmarshalmsg(buf []byte) {
    msg = detector.Msg_t{}
    err = json.Unmarshal(bytes.Trim(buf, "\x00", &msg))
    if err != nil {
        log.Fatal(err)
        return nil
    }
    return msg
}

func handlejoinreqmsg(msg Msg_t) {
    if isintroducer {
        hash := mem_table.Get_avail_hash()
        neighbors := mem_table.Get_neighbors(introducer_hash)

        // add the node to the introducers table
        mem_table.Add_node(hash, msg.Node_id)

        // send a join message to 2 previous and next nodes
        // TODO send this new node its membership list
        for i := 0; n <= len(neighbors); i++ {
            neighbor_id := mem_table.Get_node_id(neighbors[i])
            // Node id is generated in the msg
            mesg := detector.Msg_t{detector.JOIN, time.Now().UnixNano, msg.Node_id, time_to_live, hash}
            sendmessage(mesg, neighbor_id.IPV4_addr, portNum)
        }
    } else {
        log .Fatal("Only introducer should receive JOIN_REQUESTS. Ignoring")
    }
}

// a function that returns the hash of a structs members except the time to live
func hashmsgstruct(msg Msg_t) {
    s := string(msg.Msg_type) + string(msg.Timestamp) + string(Node_id.Timestamp) + string(Node_id.IPV4_addr) + string(Node_hash)
    h := fnv.New32a()
    h.Write([]byte(s))
    return h.Sum32()
}

func handlejoinmsg(msg Msg_t) {
    hash_msg := hashmsgstruct(msg)
    _, exists = message_hashes[hash_msg]
    if !exists {

        // add the node to the table
        mem_table.Add_node(msg.node_hash, msg.Node_id)

        // add it to the map, and then process it
        message_hashes[hash_msg] = 0

        // dont continue to send if no more jumps left
        if msg.Time_to_live <= 0 {
            return
        }

        msg.Time_to_live -= 1

        neighbors := mem_table.Get_neighbors(node_hash)
        for i := 0; i <= len(neighbors); i++ {
            neighbor_id := mem_table.Get_node_id(neighbors[i])
            sendmessage(msg, neighbor_id.IPV4_addr, portNum)
        }
    }
}

//TODO
func handleheartbeatmsg(msg Msg_t) {
}

func handlefailmsg(msg Msg_t) {

}

func handleleavemsg(msg Msg_t) {
    hash_msg := hashmsgstruct(msg)
    _, exists = message_hashes[hash_msg]
    if !exists {

        // delete the node from table
        mem_table.Delete_node(msg.node_hash, msg.node_id)

        message_hashes[hash_msg] = 0

        if msg.Time_to_live <= 0 {
            return
        }

        msg.Time_to_live -= 1

        neighbors := mem_table.Get_neighbors(node_hash)
        for i := 0; i <= len(neighbors); i++ {
            neighbor_id := mem_table.Get_node_id(neighbors[i])
            sendmessage(msg, neighbor_id.IPV4_addr, portNum)
        }
    }
}

func handleconnection(conn *net.UDPConn) {
    buffer := make([]byte, 1024)
    n, addr, err := conn.ReadFromUDP(buffer)
    if err != nil {
        log.Fatal(err)
    }

    msg := unmarshalmsg(buffer)
    if msg == nil {
        return nil
    }

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
            handlejoinreq(msg)
        default:
            handlemisc(msg)
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
    fmt.Println("UDP server up and listening on port " + portNum)
    defer ln.Close()
    for {
    // wait for UDP client to connect
        go handleconnection(ln)
    }
}

func init() {
    n_id := detector.Gen_node_id()
    if n_id.IPV4_addr == introducer_ip {
        isintroducer = true
        node_id = introducer_hash
    }
}

func main() {
    init()
    go listener()
}


