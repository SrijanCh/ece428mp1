package detector
import (
		 "time"
		 "fmt"
         "net"
         // "os"
   //       "strings"
		 // // "fmt"
		 // "bytes"
  //   	"os/exec"
  //   	"os"
	 //    // "log"
	 //    "fmt"
	 //    "strings"
		// "strconv"
		)

//A gob outline for the args we need for grep
// type Args struct {
	// Data, Filepath string
// }

// type Detector int; //an alias for the type we need to handle this RPC

// //The standard grep call function; takes arguments, puts them after grep, then puts the output of the grep in the reply pointer
// //Method of Querier (which we need to do because RPC)
// // Args = the struct of arguments to pass to the Grep
// // 			{ 
// // 				Data = the arguments for the grep, with flags and everything
// // 				File = the file we want to search
// // 			}
// // reply = where we'll store our resulting string
// func (t *Detector) Add_to_List(args Args, reply *string) error {

	
// 	return nil
// }

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

func Gen_msg(msg_type byte, timestamp int64, node_id Node_id_t, node_hash byte) Msg_t{
	return Msg_t{msg_type, timestamp, node_id, node_hash}
}
