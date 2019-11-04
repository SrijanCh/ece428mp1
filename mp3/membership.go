package main

import (
	"strconv"
	"fmt"
	"time"
	"net"
	"os"
	"sync"
	"strings"
	"flag"
	"reflect"
	"math/rand"
    "os/signal"
    "syscall"
    "log"
    "io"
	"sdfsrpc"
  	// "math/rand"
  	// "time"
  	// "fmt"
	"net/rpc"
	"net/http"
)

type MemberNode struct {
	ip string
	timestamp int64
	alive bool
}

type MonitorNode struct {
	conn *net.UDPConn
	ip string
	vid int
}

type ChildNode struct {
	timestamp int64
}

var introducer = "172.22.154.255"
var introducerPort = 8082
var introPingPeriod = 5

//172.22.156.255
//172.22.153.4
//172.22.155.0

var zoo_ip = "172.22.154.255"
var zoo_portnum = "3075"
var  node_portnum = "3074"

var myVid int
var otherPort = 8081
var myIP string

var memberMap = make(map[int]*MemberNode)
var monitors = make(map[string]*MonitorNode)
var children = make(map[int]*ChildNode)

var eventTimeMap = make(map[int]int64)
var fingerTable = make(map[int]int)
var fingerTablePeriod int64 = 10

var heartbeatPort = 8080
var heartbeatPeriod int64 = 1
var suspects []int 

var garbage = make(map[int]bool)

var maxID = 0
var delimiter = ","

var packetDropProb float64 = 0.0
var num_tries int = 3

func sendHeartbeat() {
	for {
		for type_, node := range(monitors) {
			if rand.Float64() > packetDropProb {
				_, err := node.conn.Write([]byte(strconv.Itoa(myVid)))
				if err != nil {
					log.Printf("[ME %d] Could not send heartbeat to %s %d", myVid, type_, node.vid)
				}
			} else {
				log.Printf("[ME %d] Dropped the heartbeat to monitor=%d", myVid, node.vid)
			}
			
			// log.Printf("[ME %d] Sent heartbeat to %s %d", myVid, type_, node.vid)
		}
		time.Sleep(time.Duration(heartbeatPeriod) * time.Second)
	}
}

func receiveHeartbeat() {
	var myaddr net.UDPAddr
	myaddr.IP = net.ParseIP(myIP)
	myaddr.Port = heartbeatPort

	heartbeatConn, err := net.ListenUDP("udp", &myaddr)
	if err != nil {
		log.Printf("Unable to setup listen on the heartbeat port %d", heartbeatPort)
	}

	log.Printf("Listening on heartbeat port %d", heartbeatPort)

	for {
		var buf [512]byte
		n, addr, err := heartbeatConn.ReadFromUDP(buf[0:])
		if err != nil {
			log.Printf("Could not read message on heartbeat port %d", heartbeatPort)
		}
		message := string(buf[0:n])
		child_vid, err := strconv.Atoi(message)
		if err != nil {
			log.Printf("Could not map a heartbeat message %s to a virtual ID\n", message)
		}

		// Check if the sender vid is in your children map
		_, ok := children[child_vid]
		if ok {
			children[child_vid].timestamp = time.Now().Unix()
			// log.Printf("[HEARTBEAT %d] Received heartbeat from child vid=%d, ip=%s\n", myVid, child_vid, addr.IP.String())
		} else{
			log.Printf("Received a non-child heartbeat from vid=%d, ip=%s", child_vid, addr.IP.String())

		}
	}
}

func printMembershipList() {
	log.Printf("My members: [")
	for id := range(memberMap) {
		log.Printf("%d ", id)
	}
	log.Printf("]\n")
}

func printChildren() {
	child_list := []int{}
	for child_vid := range children {
		child_list = append(child_list, child_vid)
	}
	log.Printf("[ME %d] Children = %v", myVid, child_list)
}

func checkChildren() {
	for {
		currTime := time.Now().Unix()

		for child_vid, cnode := range children {
			if currTime - cnode.timestamp > 2 * heartbeatPeriod {
				log.Printf("[ME %d] No heartbeat from %d since two heartbeat periods", myVid, child_vid)
				suspects = append(suspects, child_vid)
				go checkSuspicion(child_vid)
			}
		}
		time.Sleep(time.Duration(heartbeatPeriod) * time.Second)
	}
	return
}

func checkSuspicion(vid int) {
	// To check suspicion, query its neighbors
	pred := getPredecessor(vid)
	succ1 := getSuccessor(vid)
	succ2 := getSuccessor2(vid)

	for _, nvid := range([]int{pred, succ1, succ2}) {
		if nvid == myVid {
			continue
		}

		log.Printf("[ME %d] Raising suspicion for %d", myVid, vid)

		message := fmt.Sprintf("SUSPECT,%d", vid)
		sendMessage(nvid, message, num_tries)
	}

	// after 1 second, if the vid is still in suspects, declare it CRASHed
	time.Sleep(time.Duration(1) * time.Second) 

	suspect_idx := -1
	for i, suspect := range(suspects) {
		if suspect == vid {
			suspect_idx = i

			//OUR CODE
			var ip_addr = memberMap[suspect].ip
			//OUR CODE END

			memberMap[suspect].alive = false
			_, ok := children[vid]
			if ok {
				delete(children, vid)
			}
			log.Printf("[ME %d] Marked %d as crashed", myVid, suspect)
			crash_time := time.Now().Unix()
			message := fmt.Sprintf("CRASH,%d,%d", vid, crash_time)
			eventTimeMap[vid] = crash_time
			disseminate(message)
			updateMonitors()

			//PUT OUR REPLICATE HERE
			if(is_zookeeper()){
				handle_fail(ip_addr)
			}
			break
		}
	}

	// remove dead node from suspects
	if suspect_idx != -1 {
		suspects[suspect_idx] = suspects[len(suspects)-1]
		suspects = suspects[:len(suspects)-1]
	}
	return
}

func sendMessage(vid int, message string, num_tries int) {
	var addr net.UDPAddr
	addr.IP = net.ParseIP(memberMap[vid].ip)
	addr.Port = otherPort

	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		log.Printf("[ME %d] Unable to dial UDP to vid=%d ip=%s", myVid, vid, memberMap[vid].ip)
	}
	defer conn.Close()
	for i:=0; i<num_tries; i++ {
		if rand.Float64() > packetDropProb {
			_, err = conn.Write([]byte(message))
			if err != nil {
				log.Printf("[ME %d] Unable to write message %s on the connection to vid=%d", myVid, message, vid)
			}
		} else {
			log.Printf("[ME %d] Dropped the message %s to vid=%d", myVid, message, vid)
		}	
	}
	return
}

func sendMessageAddr(ip string, message string, num_tries int) {
	var addr net.UDPAddr
	addr.IP = net.ParseIP(ip)
	addr.Port = otherPort

	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		log.Printf("[ME %d] Unable to dial UDP to ip=%s", myVid, ip)
	}
	defer conn.Close()
	for i:=0; i<num_tries; i++ {
		if rand.Float64() > packetDropProb {
			_, err = conn.Write([]byte(message))
			if err != nil {
				log.Printf("[ME %d] Unable to write message %s on the connection to ip=%s", myVid, message, ip)
			}
		} else {
			log.Printf("[ME %d] Dropped the message %s to ip=%s", myVid, message, ip)
		}
	}
	return
}

// golang % operator can return negative values, define a positive mod function
func mod(a int, b int) int {
    m := a % b
    if a < 0 && b > 0 {
        m += b
    }
    return m
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
	
}

func getPredecessor(vid int) (int) {
	n := maxID + 1
	if n < 2 {
		return -1
	}

	pred := mod(vid - 1, n)
	attempts := 0
	for {
		_, ok := memberMap[pred]
		if ok && memberMap[pred].alive == true {
			if pred != vid {
				break
			}
		}
		pred = mod(pred - 1, n)
		attempts += 1
		if attempts > n {
			return -1
		}
	}
	return pred
}

func getSuccessor(vid int) (int) {
	n := maxID + 1
	if n < 2 {
		return -1
	}
	succ := (vid + 1) % n
	attempts := 0
	for {
		_, ok := memberMap[succ] // checking if succ is in the memberMap
		if ok && memberMap[succ].alive == true {
			if succ != vid {
				break
			}
		}
		succ = (succ + 1) % n
		attempts += 1
		if attempts > n {
			return -1
		}
	}
	return succ
}

func getSuccessor2(vid int) (int) {
	n := maxID + 1
	if n < 2 {
		return -1
	}

	succ1 := getSuccessor(vid)
	if succ1 == -1 {
		return -1
	}

	succ2 := (succ1 + 1) % n
	attempts := 0
	for {
		_, ok := memberMap[succ2]
		if ok && memberMap[succ2].alive == true {
			if succ2 != vid {
				break
			}
		}
		succ2 = (succ2 + 1) % n
		attempts += 1
		if attempts > n {
			return -1
		}
	}
	return succ2
}

func printFingerTable() {
	finger_list := []int{}
	for _, finger := range(fingerTable) {
		finger_list = append(finger_list, finger)
	}
	log.Printf("[ME %d] Finger Table entries = %v", myVid, finger_list)
}

func updateFingerTable() {
	for {
		n := maxID + 1
		factor := 1
		idx := 0
		for {
			if (n < 2) {
				break
			}
			val := (myVid + factor) % n
			entry := getSuccessor(val)
			if (entry != myVid) {
				fingerTable[idx] = entry
				idx = idx + 1
			}
			factor = factor * 2
			if factor >= n {
				break
			}
		}

		stale_idx := len(fingerTable)
		for i:=idx; i<stale_idx; i++ {
			_, ok := fingerTable[i]
			if ok {
				delete(fingerTable, i)
			}
		}
		printFingerTable()
		time.Sleep(time.Duration(fingerTablePeriod) * time.Second)
	}
}

func disseminate(message string) {
	for _, node := range(monitors) {
		sendMessage(node.vid, message, 1)
	}
	for _, finger := range(fingerTable) {
		if (finger == myVid || memberMap[finger].alive == false) {
			continue
		}
		sendMessage(finger, message, 1)
	}
}


func checkIntroducer() {
	for {
		time.Sleep(time.Duration(introPingPeriod) * time.Second)
		if memberMap[0].alive == false {
			// If introducer is dead, periodically send your record to the introducer
			message := fmt.Sprintf("INTRODUCER,%d,%s,%d,%d",myVid,memberMap[myVid].ip,memberMap[myVid].timestamp,maxID)
			sendMessage(0, message, num_tries)
		}
	
	}
}

func findAndSendMonitors(vid int) {
	var pred, succ1, succ2 int

	pred = getPredecessor(vid)
	if pred != -1{
		message := fmt.Sprintf("PRED,%d,%s,%d", pred, memberMap[pred].ip, memberMap[pred].timestamp)
		sendMessage(vid, message, num_tries)
	}
	
	succ1 = getSuccessor(vid)
	if succ1 != -1 {
		message := fmt.Sprintf("SUCC1,%d,%s,%d", succ1, memberMap[succ1].ip, memberMap[succ1].timestamp)
		sendMessage(vid, message, num_tries)
	}
	
	succ2 = getSuccessor2(vid)
	if succ2 != -1 {
		message := fmt.Sprintf("SUCC2,%d,%s,%d", succ2, memberMap[succ2].ip, memberMap[succ2].timestamp)
		sendMessage(vid, message, num_tries)
	}
	
}

func completeJoinRequests() (err error) {

	var myaddr net.UDPAddr
	myaddr.IP = net.ParseIP(introducer)
	myaddr.Port = introducerPort

	introducerConn, err := net.ListenUDP("udp", &myaddr)
	if err != nil {
		log.Printf("[ME %d] Unable to setup listen on the introducer port %d", myVid, introducerPort)
		return err
	}

	log.Printf("[ME %d] Started listening on the introducer port %d", myVid, introducerPort)

	for {
		var buf [512]byte
		_, addr, err := introducerConn.ReadFromUDP(buf[0:])
		if err != nil {
			log.Printf("[ME %d] Could not read message from the introducer port %d", myVid, introducerPort)
		}

		log.Printf("[ME %d] Received a JOIN request from ip=%s", myVid, addr.IP.String())

		var newVid int

		if len(garbage) == 0 {
			maxID = maxID + 1
			newVid = maxID
		} else {
			for key := range(garbage) {
				newVid = key
				break
			}
			delete(garbage, newVid)
		}

		var newnode MemberNode
		newnode.ip = addr.IP.String()
		newnode.timestamp = time.Now().Unix()
		newnode.alive = true
		memberMap[newVid] = &newnode

		log.Printf("[ME %d] Added entry ip=%s timestamp=%d at vid=%d", myVid, newnode.ip, newnode.timestamp, newVid)

		// Send the node's record
		message := fmt.Sprintf("YOU,%d,%s,%d", newVid, newnode.ip, newnode.timestamp)
		sendMessage(newVid, message, num_tries)

		// Send introducer record
		message = fmt.Sprintf("MEMBER,0,%s,%d", introducer, memberMap[0].timestamp)
		sendMessage(newVid, message, num_tries)

		findAndSendMonitors(newVid)
		
		// this delay is essential, otherwise it will be bombarded with MEMBER messages even before init setup
		time.Sleep(100 * time.Millisecond)

		message = fmt.Sprintf("JOIN,%d,%s,%d", newVid, newnode.ip, newnode.timestamp)
		eventTimeMap[newVid] = newnode.timestamp
		disseminate(message)

		updateMonitors()
	}
	return nil
}

func createMonitor(vid int) (MonitorNode) {
	var node MonitorNode
	node.vid = vid
	node.ip = memberMap[vid].ip

	var addr net.UDPAddr
	addr.IP = net.ParseIP(node.ip)
	addr.Port = heartbeatPort
	var err error
	node.conn, err = net.DialUDP("udp", nil, &addr)
	if err != nil {
		log.Printf("[ME %d] Unable to setup a connection with monitor %d", myVid, vid)
	}
	return node
}

func createMember(ip string, str_timestamp string) (MemberNode){
	var node MemberNode
	node.ip = ip
	var err error
	node.timestamp, err = strconv.ParseInt(string(str_timestamp), 10, 64)
	if err != nil {
		log.Printf("[ME %d] Cannot convert string timestamp to int64", myVid)
	}
	node.alive = true
	return node
}

// utility function for difference of lists - print when monitors updated
func Difference(a, b []int) (diff []int) {
	m := make(map[int]bool)

	for _, item := range b {
		m[item] = true
	}

    for _, item := range(a) {
    	_, ok := m[item]
    	if !ok {
    		diff = append(diff, item)
    	}
    }

    return
}

func updateMonitors() {

	old_monitors := []int{}
	new_monitors := []int{}

	_, ok := monitors["pred"]
	newpred := getPredecessor(myVid)

	if ok {
		old_monitors = append(old_monitors, monitors["pred"].vid)
		delete(monitors, "pred")
	}
	if newpred != -1 {
		monitor_node := createMonitor(newpred)
		monitors["pred"] = &monitor_node
		new_monitors = append(new_monitors, newpred)
	}

	_, ok = monitors["succ1"]
	newsucc1 := getSuccessor(myVid)

	if ok {
		old_monitors = append(old_monitors, monitors["succ1"].vid)
		delete(monitors, "succ1")
	}
	if newsucc1 != -1 {
		monitor_node := createMonitor(newsucc1)
		monitors["succ1"] = &monitor_node
		new_monitors = append(new_monitors, newsucc1)
	}

	_, ok = monitors["succ2"]
	newsucc2 := getSuccessor2(myVid)

	if ok {
		old_monitors = append(old_monitors, monitors["succ2"].vid)
		delete(monitors, "succ2")
	}
	if newsucc2 != -1 {
		monitor_node := createMonitor(newsucc2)
		monitors["succ2"] = &monitor_node
		new_monitors = append(new_monitors, newsucc2)
	}

	to_add := Difference(new_monitors, old_monitors)
	for _, vid := range(to_add) {
		message := fmt.Sprintf("ADD,%d,%s,%d", myVid, memberMap[myVid].ip, memberMap[myVid].timestamp)
		sendMessage(vid, message, num_tries)
	}

	to_remove := Difference(old_monitors, new_monitors)
	for _, vid := range(to_remove) {
		message := fmt.Sprintf("REMOVE,%d", myVid)
		sendMessage(vid, message, num_tries)
	}

	if !reflect.DeepEqual(old_monitors, new_monitors) {
		log.Printf("[ME %d] Updated monitors from %v to %v", myVid, old_monitors, new_monitors)
	}
}

func printGarbage() {
	garbage_list := []int{}
	for k := range(garbage) {
		garbage_list = append(garbage_list, k)
	}
	log.Printf("[ME %d] Garbage set = %v", myVid, garbage_list)
}


func garbageCollection() {
	for {
		time.Sleep(30 * time.Second)

		for i:=1; i<=maxID; i++ {
			mnode, isavailable := memberMap[i]
			if ((!isavailable || !mnode.alive) && (time.Now().Unix() - eventTimeMap[i] > 6)) {
				garbage[i] = true
			}
		}
		printGarbage()
	}
}


func listenOtherPort() (err error) {
	var myaddr net.UDPAddr
	myaddr.IP = net.ParseIP(myIP)
	myaddr.Port = otherPort

	otherportConn, err := net.ListenUDP("udp", &myaddr)
	log.Printf("[ME %d] Started listening on the otherport %d", myVid, otherPort)

	if err != nil {
		log.Printf("[ME %d] Unable to listen on the otherport port %d", myVid, otherPort)
		return err
	}

	for {
		var buf [512]byte
		n, addr, err := otherportConn.ReadFromUDP(buf[0:])
		if err != nil {
			log.Printf("[ME %d] Could not read message on otherport %s", myVid, otherPort)
		}

		message := string(buf[0:n])
		split_message := strings.Split(message, delimiter)
		message_type := split_message[0]
		subject, _ := strconv.Atoi(split_message[1])
		maxID = max(maxID, subject)

		log.Printf("[ME %d] Message = %s", myVid, message)

		switch message_type {
		case "ADD":
			var newnode MemberNode
			newnode = createMember(split_message[2], split_message[3])
			memberMap[subject] = &newnode

			var cnode ChildNode
			cnode.timestamp = time.Now().Unix()
			children[subject] = &cnode

			printChildren()

		case "REMOVE":
			_, ok := children[subject]
			if ok {
				delete(children, subject)
			}

			printChildren()

		case "INTRODUCER":
			if myVid == 0 {
				// Listen to atleast 4 different nodes than myself - to handle three simultaneous failures
				if len(memberMap) < 5 {
					newnode := createMember(split_message[2],split_message[3])
					memberMap[subject] = &newnode

					tempmax, _ := strconv.Atoi(split_message[4])
					maxID = max(maxID, tempmax)
					
					message := fmt.Sprintf("JOIN,%d,%s,%d", 0, memberMap[0].ip,memberMap[0].timestamp)
					updateMonitors()

					eventTimeMap[0] = memberMap[0].timestamp
					disseminate(message)

					log.Printf("[ME %d] Processed introducer ping entry from vid=%d", myVid, subject)
				}
			}

		case "PRED", "SUCC1", "SUCC2":
			var newnode MemberNode
			newnode = createMember(split_message[2], split_message[3])
			memberMap[subject] = &newnode

			var node MonitorNode
			node = createMonitor(subject)
			monitors[strings.ToLower(message_type)] = &node

			old_mon, ok := monitors[strings.ToLower(message_type)]
			if ok {
				old_mon_vid := old_mon.vid
				if old_mon_vid != subject {
					message := fmt.Sprintf("REMOVE,%d", myVid)
					sendMessage(old_mon_vid, message, num_tries)
				}
			}

			message := fmt.Sprintf("ADD,%d,%s,%d", myVid, memberMap[myVid].ip, memberMap[myVid].timestamp)
			sendMessageAddr(newnode.ip, message, num_tries)

			log.Printf("[ME %d] Set my %s to %d", myVid, strings.ToLower(message_type), subject)

		case "YOU":
			myVid = subject
			var newnode MemberNode
			newnode = createMember(split_message[2], split_message[3])
			memberMap[subject] = &newnode

			go checkIntroducer()

			log.Printf("[ME %d] Processed my memberMap entry", myVid)

		case "MEMBER":
			if subject == myVid {
				break
			}
			newnode := createMember(split_message[2], split_message[3])
			memberMap[subject] = &newnode

			if subject != 0 {
				updateMonitors()
				// introducer is anyway going to send it monitors.
			}
			
			log.Printf("[ME %d] Processed a new memberMap entry vid=%d", myVid, subject)

		case "JOIN":
			origin_time, _ := strconv.ParseInt(string(split_message[3]), 10, 64)

			_, ok := eventTimeMap[subject]
			if (!ok || eventTimeMap[subject] < origin_time) {
				eventTimeMap[subject] = origin_time
				disseminate(message)

				if subject != myVid {
					newnode := createMember(split_message[2], split_message[3])
					memberMap[subject] = &newnode
				}

				message := fmt.Sprintf("MEMBER,%d,%s,%d", myVid, myIP, memberMap[myVid].timestamp)
				sendMessage(subject, message, num_tries)

				updateMonitors()

				log.Printf("[ME %d] Processed JOIN memberMap entry for vid=%d", myVid, subject)

				printMembershipList()
			} 

		case "LEAVE", "CRASH":
			origin_time, _ := strconv.ParseInt(string(split_message[2]), 10, 64)

			_, ok := eventTimeMap[subject]
			if (!ok || eventTimeMap[subject] < origin_time){
				eventTimeMap[subject] = origin_time
				disseminate(message)

				_, ok := memberMap[subject]
				if ok {

					//OUR CODE
					var ip_addr = memberMap[subject].ip
					//OUR CODE END

					memberMap[subject].alive = false
					
					_, ok = children[subject]
					if ok {
						delete(children, subject)
					}

					if subject == maxID {
						var i int
						for i=maxID; i>=0; i-- {
							if memberMap[i].alive {
								break
							}
						}
						maxID = i						
					}

					updateMonitors()

					//PUT OUR REPLICATE HERE
					if(is_zookeeper()){
						handle_fail(ip_addr)
					}
					
					log.Printf("[ME %d] Processed %s for %d, maxID = %d", myVid, message_type, subject, maxID)
				}
			}			

		case "SUSPECT":
			var alive = false

			// Checked if it is set as dead in my list, if yes send dead message already
			if memberMap[subject].alive == false {
				alive = false
			} else{
				var currTime = time.Now().Unix()
				for child_vid, cnode := range(children) {
					if subject == child_vid {
						if currTime - cnode.timestamp < heartbeatPeriod {
							alive = true
						}
						break
					}
				}
			}

			var message string
			if alive {
				message = fmt.Sprintf("STATUS,%d,1", subject)
			} else {
				message = fmt.Sprintf("STATUS,%d,0", subject)
			}
			sendMessageAddr(addr.IP.String(), message, num_tries)
			if alive {
				log.Printf("[ME %d] Processed suspect message for %d, sent ALIVE", myVid, subject)
			} else {
				log.Printf("[ME %d] Processed a suspect message for %d, sent NOT ALIVE", myVid, subject)
			}
		
		case "STATUS":
			status, _ := strconv.Atoi(split_message[2])
			if status == 1 {
				suspect_idx := -1
				for i, suspect := range(suspects) {
					if suspect == subject {
						suspect_idx = i
						break
					}
				}
				if suspect_idx != -1 {
					suspects[suspect_idx] = suspects[len(suspects)-1]
					suspects = suspects[:len(suspects)-1]
				}
			}
			if status == 1 {
				log.Printf("[ME %d] Received ALIVE status for %d", myVid, subject)
			} else {
				log.Printf("[ME %d] Received NOT ALIVE status for %d", myVid, subject)
			}
			
		}
	}
}

func sendJoinRequest() {
	var addr net.UDPAddr
	addr.IP = net.ParseIP(introducer)
	addr.Port = introducerPort

	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		log.Printf("[ME %d] Unable to dial UDP to introducer ip=%s, port=%d", myVid, introducer, introducerPort)
	}
	message := "1"
	defer conn.Close()
	conn.Write([]byte(message))

	log.Printf("[ME %d] Sent a JOIN request to introducer ip=%s", myVid, introducer)
	return
}

func getmyIP() (string) {
	var myip string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalf("Cannot get my IP")
		os.Exit(1)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				myip = ipnet.IP.String()
			}
		}
	}
	return myip
}

func main() {
	logFile := flag.String("logfile", "", "path to the log file")
	flag.Parse()

	var wg sync.WaitGroup
	wg.Add(1)

	f, err := os.Create(*logFile)
	if err != nil {
		log.Fatalf("Error opening the log file: %v", err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	myIP = getmyIP()
	log.Printf("%s", myIP)

	if myIP == introducer {
		myVid = 0
		var node MemberNode
		node.ip = myIP
		node.timestamp = time.Now().Unix()
		node.alive = true
		memberMap[0] = &node
	}

	go sendHeartbeat()
	go receiveHeartbeat()
	go checkChildren()

	go listenOtherPort()

	time.Sleep(time.Duration(introPingPeriod) * time.Second)
	
	if is_zookeeper(){//myIP == introducer {
		// there should be a delay here - depending on how frequently the introducer is being pinged
		// if the system already exists in some form, the introducer shouldn't accept join requests until it knows what the maxID is 
		go completeJoinRequests()
		go garbageCollection()
		// this garbage collection can occur concurrent to the addToDead list

	} else{
		sendJoinRequest()
	}
	go updateFingerTable()


	//OUR STUFF
		if is_zookeeper(){
			log.Printf("Starting Zookeeper\n")
			//DEPLOY ZOOKEEPER
			go host_zookeeper()
			//ZOOKEEPER END
		} else {
			//DEPLOY SDFS
			log.Printf("Starting SDFS\n")
			go host_sdfs()
			//SDFS END
		}
	//OUR STUFF END
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGQUIT)

	go func() {
		sig := <-sigs
		switch sig {
		case syscall.SIGQUIT:
			leave_time := time.Now().Unix()
			message := fmt.Sprintf("LEAVE,%d,%d", myVid, leave_time)
			disseminate(message)
			
			wg.Done()
		}
	}()


	wg.Wait()
	return
}


//////////////////////////////////////OUR FILE DUMP//////////////////////////////////////////////////
// const node_portnum = "3074"         

// type MemberNode struct {
// 	ip string
// 	Timestamp int64
// 	alive bool
// }

// var memberMap = make(map[int]*MemberNode)
// var Garbage = make(map[int]bool)
// var FailQueue = make([]*MemberNode)

type Zookeeper int

type FileLoc struct{
	MemID 	  int
	ip 		  string
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
	printmemberMap();
	count := 0
	for _, v := range memberMap{
		if v.alive{	
			count++
		}
	}
	return count
}

func printmemberMap() {
	fmt.Printf("Zookeeper member list: [\n")
	for id, ele := range(memberMap) {
		fmt.Printf("%d: %s, %d, %t\n", id, ele.ip, ele.timestamp, ele.alive)
	}
	fmt.Printf("]\n")
}

func pick4() ([]int, []string){
	var ret_str []string = make([]string, 4)
	var ret_int []int = make([]int, 4)
	var a,b,c,d int

	if (num_live() < 4){
		fmt.Printf("Not enough live nodes to pick random, %d\n", num_live())
		for k, v := range memberMap{
			if v.alive{	
				ret_str = append(ret_str, v.ip)
				ret_int = append(ret_int, k)
				fmt.Printf("Returning %s and %d in set\n", v.ip, k)
			}
		}
		return ret_int, ret_str
	}

	fmt.Printf("Picking 4 random nodes\n")

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	a = r.Intn(len(memberMap)-1)+1
	_, ok := memberMap[a]
	for !ok || !memberMap[a].alive{
		a = r.Intn(len(memberMap)-1)+1
	}
	ret_str[0] = memberMap[a].ip
	ret_int[0] = a
	
	b = r.Intn(len(memberMap)-1)+1
	_, ok = memberMap[b]
	for !ok || !memberMap[b].alive || b == a{
		b = r.Intn(len(memberMap)-1)+1
	}
	ret_str[1] = memberMap[b].ip
	ret_int[1] = b


	c = r.Intn(len(memberMap)-1)+1
	_, ok = memberMap[c]
	for !ok || !memberMap[c].alive || c == a || c == b{
		c = r.Intn(len(memberMap)-1)+1
	}
	ret_str[2] = memberMap[c].ip
	ret_int[2] = c


	d = r.Intn(len(memberMap)-1)+1
	_, ok = memberMap[d]
	for !ok || !memberMap[d].alive || d == a || d == b || d == c{
		d = r.Intn(len(memberMap)-1)+1
	}
	ret_str[3] = memberMap[d].ip
	ret_int[3] = d

	fmt.Printf("Picked [(%d,%d) %s,%s] [(%d,%d) %s,%s] [(%d,%d) %s,%s] and [(%d,%d) %s,%s]\n", a, ret_int[0], 
																							   memberMap[a].ip, ret_str[0],
																							   b, ret_int[1],
																							   memberMap[b].ip, ret_str[1],
																							   c, ret_int[2],
																							   memberMap[c].ip, ret_str[2],
																							   d, ret_int[3],
																							   memberMap[d].ip, ret_str[3])
	return ret_int, ret_str
}

func pick3(fileloc_arr [4]FileLoc) ([]int, []string){
	var ret_str []string = make([]string, 3)
	var ret_int []int = make([]int, 3)
	//Pick one to not
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	a := r.Intn(len(fileloc_arr))
	for !memberMap[fileloc_arr[a].MemID].alive{
		a = r.Intn(len(fileloc_arr))
	}
	ret_int[0] = a
	ret_str[0] = fileloc_arr[a].ip


	b := r.Intn(len(fileloc_arr))
	for !memberMap[fileloc_arr[b].MemID].alive || a==b{
		b = r.Intn(len(fileloc_arr))
	}
	ret_int[1] = b
	ret_str[1] = fileloc_arr[b].ip


	c := r.Intn(len(fileloc_arr))
	for !memberMap[fileloc_arr[c].MemID].alive || c==a || c==b{
		c = r.Intn(len(fileloc_arr))
	}
	ret_int[2] = c
	ret_str[2] = fileloc_arr[c].ip


	fmt.Printf("Picked [(%d, MemID %d) %s,%s] [(%d, MemID %d) %s,%s] [(%d, MemID %d) %s,%s]\n",  			
																			a, fileloc_arr[a].MemID,
																		    fileloc_arr[a].ip, ret_str[0],
																		    b, fileloc_arr[b].MemID,
																		    fileloc_arr[b].ip, ret_str[1],
																			c, fileloc_arr[c].MemID,
																			fileloc_arr[c].ip, ret_str[2])
	return ret_int, ret_str
}

func pick2(fileloc_arr [4]FileLoc) ([]int, []string){
	var ret_str []string = make([]string, 2)
	var ret_int []int = make([]int, 2)
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	a := r.Intn(len(fileloc_arr))
	for !memberMap[fileloc_arr[a].MemID].alive{
		a = r.Intn(len(fileloc_arr))
	}
	ret_int[0] = a
	ret_str[0] = fileloc_arr[a].ip


	b := r.Intn(len(fileloc_arr))
	for !memberMap[fileloc_arr[b].MemID].alive || a==b{
		b = r.Intn(len(fileloc_arr))
	}
	ret_int[1] = b
	ret_str[1] = fileloc_arr[b].ip

	
	fmt.Printf("Picked [(%d, MemID %d) %s,%s] [(%d, MemID %d) %s,%s TIMESTAMPS %d, %d]\n", 	a, fileloc_arr[a].MemID,
																		    fileloc_arr[a].ip, ret_str[0],
																		    b, fileloc_arr[b].MemID,
																		    fileloc_arr[b].ip, ret_str[1],
																			fileloc_arr[a].Timestamp, fileloc_arr[b].Timestamp)
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
		fmt.Println("a: ", a)
		fmt.Println("b: ", b)
		c := int64(time.Now().Unix())
		fmt.Printf("PUT: THIS TIMESTAMP IS %d\n", c)
		var f [4]FileLoc
		f[0] = FileLoc{a[0], b[0], c}
		f[1] = FileLoc{a[1], b[1], c}
		f[2] = FileLoc{a[2], b[2], c}
		f[3] = FileLoc{a[3], b[3], c}
		FileTable[args.Sdfsname] = f
		(*reply).Ips = b
		fmt.Println("(*reply).Ips: ", (*reply).Ips)
		(*reply).Timestamp = c
	}else{									   //UPDATE (QUORUM)
		//Check on timestamp to see when last write was
		//Pick random 3 because Quorum
		a, b := pick3(FileTable[args.Sdfsname])
		c := int64(time.Now().Unix())
		fmt.Printf("PUT: THIS (REWRITE) TIMESTAMP IS %d\n", c)
		var f [4]FileLoc
		miss := 6 - a[0] - a[1] - a[2]
		f[a[0]] = FileLoc{fileloc_arr[a[0]].MemID, fileloc_arr[a[0]].ip, c}
		f[a[1]] = FileLoc{fileloc_arr[a[1]].MemID, fileloc_arr[a[1]].ip, c}
		f[a[2]] = FileLoc{fileloc_arr[a[2]].MemID, fileloc_arr[a[2]].ip, c}
		f[miss] = FileLoc{fileloc_arr[miss].MemID, fileloc_arr[miss].ip, fileloc_arr[miss].Timestamp}
		FileTable[args.Sdfsname] = f

		fmt.Println("The new table: ", FileTable[args.Sdfsname])

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
		a, b := pick2(fileloc_arr)
		c0 := ((FileTable[args.Sdfsname])[a[0]]).Timestamp//get_timestamp(args.Sdfsname, b[0], node_portnum)
		c1 := ((FileTable[args.Sdfsname])[a[1]]).Timestamp//get_timestamp(args.Sdfsname, b[1], node_portnum)
		if(c0 == c1){ 		//both are on same consistency
			fmt.Printf("Equal timestamps!\n")
			(*reply).Ip = b[0]
		} else {				//One needs an update
			if c0 > c1 {
				fmt.Printf("c0 is more recent\n")
				(*reply).Ip = b[0]
				rep_to(args.Sdfsname, b[1], node_portnum, b[0], node_portnum)
			}else{
				fmt.Printf("c1 is more recent\n")
				(*reply).Ip = b[1]
				rep_to(args.Sdfsname, b[0], node_portnum, b[1], node_portnum)
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
			del(args.Sdfsname, (FileTable[args.Sdfsname])[i].ip, node_portnum)
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
			str += "(" + string((FileTable[args.Sdfsname])[i].MemID) + ")" + string((FileTable[args.Sdfsname])[i].ip) + "\n"
		}
	}
	*reply = str
	return nil

}
// 5) store: At any machine, list all files currently being stored at this
// machine. Here sdfsfilename is an arbitrary string while localfilename is
// the Unix-style local file system name.
type Store_args struct {
    Node_ip string
}

func (t* Zookeeper) Zoo_store(args Store_args, reply *string) error {
    *reply = store(args.Node_ip, node_portnum)
    return nil
}


type Table_args struct{
	table map[string]([4]FileLoc)
}
func (t* Zookeeper) Zoo_update_table(args Table_args, reply *string) error {
    fmt.Println(args.table)
    FileTable = args.table
    *reply = ""
    return nil
}

// Gather all the files stored at the specified ip address
func store(ip, port string) string {
	fmt.Printf("Get_store, ip: %s, port %s...\n", ip, port)
    client, err := rpc.DialHTTP("tcp", ip + ":" + port)
    if err != nil {
    	fmt.Printf("Error setting up RPC dial...\n")
        return ""
    }
    var args = 0//sdfsrpc.Read_args{Sdfsname: ip}
    var reply string
    err = client.Call("Sdfsrpc.Get_store", args, &reply)
    if err != nil {
    	fmt.Printf("Error with RPC call...\n")
        return ""
    }
	fmt.Printf("Got reply %s...\n", reply)
    return reply
}

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

func host_zookeeper(){
	zookeeper := new(Zookeeper) //Creates a new Zookeeper object to handle the RPCs for this server
	rpc.Register(zookeeper) //Registers the Zookeeper as our handler
	rpc.HandleHTTP() //HTTP format requests
	fmt.Printf("Zookeeper is up for business! \n"); //Notify user
	l, e := net.Listen("tcp", ":" + zoo_portnum) //Listen to requests on port 3075
	if e != nil {
		//Error handling
		log.Fatal("listen error:", e)
	}
	fmt.Printf("Serving on port %s\n", zoo_portnum); //Notify that we're up
	http.Serve(l, nil) //Serve	
}

func host_sdfs(){
	sdfsrpc := new(sdfsrpc.Sdfsrpc) //Creates a new Querier object to handle the RPCs for this server
	rpc.Register(sdfsrpc) //Registers the Querier as our handler
	rpc.HandleHTTP() //HTTP format requests
	fmt.Printf("Sdfsrpc server start listening: \n"); //Notify user
	l, e := net.Listen("tcp", ":" + node_portnum)//3074") //Listen to requests on port 3074
	if e != nil {
		//Error handling
		log.Fatal("listen error:", e)
	}
	fmt.Printf("Serving RPC server on port %d\n", 3074); //Notify that we're up
	http.Serve(l, nil) //Serve
}

func is_zookeeper() bool{
	return myIP == zoo_ip
}




//~~~~~~~~~~~~~~~~~~~~~~~~~~~~NODE FAILURE HANDLING~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
func handle_fail(ip_addr string){
	fmt.Printf("Detected %s failed, re-replicating...\n")
	file_vec := find_assoc_files(ip_addr)
	for _,file := range file_vec{
		fmt.Printf("Replicating file %s\n", file)
		rel_ip := find_reliable_replica(file, ip_addr)
		new_mem_id, new_ip := find_new_replica(file, ip_addr)
		fmt.Printf("Replicating from %s:%s to %s:%s...\n", rel_ip, node_portnum, new_ip, node_portnum)
		rep_to(file, new_ip, node_portnum, rel_ip, node_portnum)
		fmt.Printf("Replication done; updating table...")
		var f [4]FileLoc
		for j := 0; j < 4; j++ {
			if (FileTable[file])[j].ip != ip_addr{
				f[j] = FileLoc{(FileTable[file])[j].MemID, (FileTable[file])[j].ip, (FileTable[file])[j].Timestamp}
			} else {
				f[j] = FileLoc{new_mem_id, new_ip, int64(time.Now().Unix())}
			}
		}
		FileTable[file] = f
		fmt.Println("New FileTable entry: ", FileTable[file])	
		fmt.Printf("Done with Replicating %s, moving on...\n", file)
	}
	fmt.Printf("Done with recovery from death of %s!\n", ip_addr)
}


// type FileLoc struct{
// 	MemID 	  int
// 	ip 		  string
// 	Timestamp int64
// }

// var FileTable = make(map[string]([4]FileLoc))

func find_assoc_files(ip_addr string) []string{
	fmt.Printf("Finding associated files for %s\n", ip_addr)
	var fvec []string
	for k, v := range FileTable{
		fmt.Printf("Checking file %s\n", k)
		fmt.Println("The list is ", v)
		for _, copy := range v{
			fmt.Printf("Checking file %s ip address %s against our dead address %s\n", k, copy.ip, ip_addr)
			if copy.ip == ip_addr{
				fmt.Printf("Found a match for %s! Adding to vector...\n", k)
				fvec = append(fvec, k)
				fmt.Println("Current file_vec: ", fvec)
			}
		}
	}
	fmt.Println("Done! fvec finalized as ", fvec)
	return fvec
}

func find_reliable_replica(file, dead_ip string) string{
	fmt.Printf("Finding a reliable file copy for %s...\n", file)
	var max_ts int64 = 0
	var max_idx int = 5
	for idx, copy := range FileTable[file]{
		if copy.ip != dead_ip && memberMap[copy.MemID].alive && copy.Timestamp > max_ts{
			fmt.Printf("So far found %s at %d as the most recent replica, with timestamp %d\n", copy.ip, idx, copy.Timestamp)
			max_ts = copy.Timestamp
			max_idx = idx
		}
	}

	if max_idx == 5{
		fmt.Printf("We somehow messed up; no replicas found\n")
	}

	fmt.Printf("Final replica of file %s: %s at %d as the most recent replica, with timestamp %d\n", file, (FileTable[file])[max_idx].ip, max_idx, max_ts)
	return (FileTable[file])[max_idx].ip 
}


// type MemberNode struct {
// 	ip string
// 	timestamp int64
// 	alive bool
// }
// var memberMap = make(map[int]*MemberNode)

func find_new_replica(file, dead_ip string) (int, string){
	fmt.Printf("Finding a new node for %s...\n", file)
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator
	a := r.Intn(len(memberMap)-1)+1
	ip := memberMap[a].ip
	fmt.Printf("Picked %s at %d\n", ip, a)
	for ip == dead_ip || ip == (FileTable[file])[0].ip || ip == (FileTable[file])[1].ip || ip == (FileTable[file])[2].ip || ip == (FileTable[file])[3].ip || !memberMap[a].alive{

		fmt.Printf("Rerolling...\n")
		a = r.Intn(len(memberMap)-1)+1
		ip = memberMap[a].ip
		fmt.Printf("Picked %s at %d\n", ip, a)
	}
	return a, ip 
}


//WHEN DOING ZOOKEEPER HOT SWAP
	//Change randomizers
	//Change zoo_ip