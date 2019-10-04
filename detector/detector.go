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

type node_id_t struct{
	timestamp int
	IPV4_addr net.IP
}

type msg_t struct{
	msg_type byte
	timestamp int
	node_id node_id_t
	node_hash byte
}

func gen_node_id() node_id_t{
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
    return node_id_t{timestamp, currentIP}
}