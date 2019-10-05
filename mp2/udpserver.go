package main

 import (
         "fmt"
         "log"
         "net"
         "detector"
         "encoding/json"
         "bytes"
 )

 var node_id = ""
 var node_hash = -1
 var mem_table = make(map[int]string) 

 func handleUDPConnection(conn *net.UDPConn) {

         // here is where you want to do stuff like read or write to client
         buffer := make([]byte, 1024)
         n, addr, err := conn.ReadFromUDP(buffer)
         fmt.Println("UDP client : ", addr)
         fmt.Println("Received from UDP client :  ", string(buffer[:n]))
         if err != nil {
                 log.Fatal(err)
         }
         // NOTE : Need to specify client address in WriteToUDP() function
         //        otherwise, you will get this error message
         //        write udp : write: destination address required if you use Write() function instead of WriteToUDP()
         incoming := &detector.Msg_t{} 
         // write message back to client
         err = json.Unmarshal(bytes.Trim(buffer, "\x00"), &incoming) // must trim the nulls otherwise unmarshalling fails
         if err != nil {
            fmt.Println("Unmarshalling failed")
         }
         fmt.Println("Server:")
         fmt.Println("Message = ", incoming.Timestamp)
         _, err = conn.WriteToUDP(buffer, addr)

         if err != nil {
                 log.Println(err)
         }

 }

type test struct {
    Val1 int
    Val2 int64
    Val3 int64
}

 func main() {
            
         t := &detector.Msg_t{detector.JOIN, 2000000, detector.Gen_node_id(), 100}
         marshal_res, err := json.Marshal(t)

         if err != nil {
            fmt.Println("Marshalling failed")
         }
         
         unmarshal_res := &detector.Msg_t{}

         err = json.Unmarshal([]byte(marshal_res), &unmarshal_res)
         if err != nil {
            fmt.Println("Unmarshalling failed")
         }
         fmt.Println("Result:")
         fmt.Println(unmarshal_res)
         
         hostName := "localhost"
         portNum := "6000"
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

         fmt.Println("UDP server up and listening on port 6000")

         defer ln.Close()

         for {
                 // wait for UDP client to connect
                 handleUDPConnection(ln)
         }

 }
