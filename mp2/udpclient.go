package main

 import (
         "log"
         "net"
         "fmt"
         "time"
         "encoding/json"
         "detector"
 )

 func main() {
         hostName := "localhost"
         portNum := "6000"
         service := hostName + ":" + portNum
         RemoteAddr, err := net.ResolveUDPAddr("udp", service)

         //LocalAddr := nil
         // see https://golang.org/pkg/net/#DialUDP

         conn, err := net.DialUDP("udp", nil, RemoteAddr)

         // note : you can use net.ResolveUDPAddr for LocalAddr as well
         //        for this tutorial simplicity sake, we will just use nil

         if err != nil {
                 log.Fatal(err)
         }

         log.Printf("Established connection to %s \n", service)
         log.Printf("Remote UDP address : %s \n", conn.RemoteAddr().String())
         log.Printf("Local UDP client address : %s \n", conn.LocalAddr().String())

         defer conn.Close()
         t := time.Now().UnixNano()
         message := &msg{JOIN, t, 0, 100}
         s, _ :=  json.Marshal(message)
         // write a message to server
         message := []byte(string(s))

         _, err = conn.Write(message)

         if err != nil {
                 log.Println(err)
         }

         // receive message from server
         buffer := make([]byte, 1024)
         n, addr, err := conn.ReadFromUDP(buffer)

         fmt.Println("UDP Server : ", addr)
         fmt.Println("Received from UDP server : ", string(buffer[:n]))

 }
