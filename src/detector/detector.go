package detector
import (
		 "time"
		 "fmt"
         "net"
		)


const(
	HEARTBEAT byte = 0
	JOIN byte = 1
	FAIL byte = 2
	LEAVE byte = 3
	JOIN_REQ byte = 4
	MISC byte = 5
)

type Node_id_t struct{
	Timestamp int
	IPV4_addr net.IP
}

type Msg_t struct{
	Msg_type byte
	Timestamp int64
	Node_id Node_id_t
    Time_to_live byte
	Node_hash byte
}

func Gen_node_id() Node_id_t{
	fmt.Println("Generating a node address for %s\n", )
	a := time.Now()
	timestamp := a.Nanosecond()

	addrs, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println(err)
    }

    var currentIP net.IP //, currentNetworkHardwareName string
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        // = GET LOCAL IP ADDRESS
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                    fmt.Println("Current IP address : ", ipnet.IP.String())
                    currentIP = ipnet.IP//.String()
            }
        }
    }
    return Node_id_t{timestamp, currentIP}
}

