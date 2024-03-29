package sdfsrpc

import (
		// "bytes"
    	// "os/exec"
	    // "log"
	    // "strings"
		// "strconv"
		// "bufio"
	    "fmt"
    	"os"
    	"io/ioutil"
		"net/rpc"
		"time"
		)

type Sdfsrpc int //an alias for the type we need to handle this RPC

var Filemap = make(map[string]int64)

///////////////////////////////////////////////////////////////////
//A gob outline for the args we need for grep
type Write_args struct {
	Sdfsname, Data string
	Timestamp int64
}

//Writes to a file anew
func (t *Sdfsrpc) Write_file(args Write_args, reply *int) error {
	// Make the file in case it does not exist yet
	fmt.Printf("---------------------------Writing %s to SDFS...------------------------\n", args.Sdfsname)

	//File I/O way to do it
	file, err := os.Create(args.Sdfsname)
	if err != nil{
		fmt.Printf("File creation error: %s\n", args.Sdfsname)
		return err
	}
	defer file.Close()
    bytesWritten, err := file.WriteString(args.Data)
	if(err != nil){
		fmt.Printf("File write error: %s\n", args.Sdfsname)
		return err
	}

	Filemap[args.Sdfsname] = args.Timestamp
	*reply = bytesWritten
	return nil
}

//Appends to the end of a file
func (t *Sdfsrpc) Append_file(args Write_args, reply *int) error {
	// Make the file in case it does not exist yet
	fmt.Printf("---------------------------Writing %s to SDFS...------------------------\n", args.Sdfsname)

	//File I/O way to do it
	file, err := os.OpenFile(args.Sdfsname, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil{
		fmt.Printf("File creation error: %s\n", args.Sdfsname)
		return err
	}
	defer file.Close()
    bytesWritten, err := file.WriteString(args.Data)
	if(err != nil){
		fmt.Printf("File write error: %s\n", args.Sdfsname)
		return err
	}

	Filemap[args.Sdfsname] = args.Timestamp
	*reply = bytesWritten
	return nil
}


////////////////////////////////////////////////////////////////////////////
type Read_args struct {
	Sdfsname string
}
type Read_reply struct {
	Data string
	Timestamp int64
}

//Gets a file by sdfsname
func (t *Sdfsrpc) Get_file(args Read_args, reply *Read_reply) error {
	fmt.Printf("---------------------------Getting %s from SDFS...---------------------------\n", args.Sdfsname)

	if ts, ok := Filemap[args.Sdfsname]; !ok {
		fmt.Printf("We don't have file\n")
		(*reply).Data = ""
		(*reply).Timestamp = 0
		return nil
	}else{
		//File I/O way to do it
		file, err := os.Open(args.Sdfsname)
		if(err != nil){
			fmt.Printf("File open error: %s\n", args.Sdfsname)
			(*reply).Data = ""
			(*reply).Timestamp = 0
			return err
		}

		defer file.Close()
    	Data_r, err := ioutil.ReadAll(file)

		if(err != nil){
			fmt.Printf("File read error: %s\n", args.Sdfsname)
			(*reply).Data = ""
			(*reply).Timestamp = 0
			return err
		}
	
		(*reply).Data = string(Data_r)
		(*reply).Timestamp = ts
		return nil
	}
}

/////////////////////////////////////////////////////////////////////

//Gets the timestamp for a file
func (t *Sdfsrpc) Get_timestamp(args Read_args, reply *int64) error {
	fmt.Printf("Getting %s's Timestamp from SDFS...\n", args.Sdfsname)
	
	if ts, ok := Filemap[args.Sdfsname]; !ok {
		fmt.Printf("We don't have file %s\n", args.Sdfsname)
		*reply = 0
		return nil
	}else{
		*reply = ts
	}
	return nil
}

/////////////////////////////////////////////////////////////////////

//Deletes a file
func (t *Sdfsrpc) Delete_file(args Read_args, reply *int64) error{
	fmt.Printf("---------------------------Deleting %s from SDFS...---------------------------\n", args.Sdfsname)

	if _, ok := Filemap[args.Sdfsname]; !ok {
		fmt.Printf("Already don't have file\n")
		*reply = 0
		return nil
	}else{
    	err := os.Remove(args.Sdfsname)
    	if err != nil {
    	    fmt.Printf("Delete error for %s\n", args.Sdfsname);
    	    *reply = 0
    	    return err
    	}
    	delete(Filemap, args.Sdfsname)
    	*reply = 1
    	fmt.Printf("Deleted %s successfully\n", args.Sdfsname)
    	return nil
    }
}

//////////////////////////////////////////////////////////////////////

//Gets a list of stored files
func (t *Sdfsrpc) Get_store(args int, reply *string) error {
	fmt.Printf("~~~~~~~~~~~~~~~~~~~~~~~~Get_store~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n")
	for k, _ := range Filemap{
		(*reply) += k + "\n"
	}
	fmt.Printf("Get_store returning %s\n", (*reply))
	return nil
}

//////////////////////////////////////////////////////////////////////
type Rep_args struct{
	Sdfsname, Ip, Port string
}

//Replicates from a file to another one
func (t *Sdfsrpc) Replicate_to(args Rep_args, reply *int) error {
		fmt.Printf("---------------------------Replicate_to %s, %s, %s---------------------------\n", args.Sdfsname, args.Ip, args.Port)
	if _, ok := Filemap[args.Sdfsname]; !ok {
		fmt.Printf("We don't have file %s\n", args.Sdfsname)
		(*reply) = 0
		return nil
	}else{
		//File I/O way to do it
		fmt.Printf("Have file! opening...\n")
		file, err := os.Open(args.Sdfsname)
		if(err != nil){
			fmt.Printf("File open error: %s\n", args.Sdfsname)
			(*reply) = 0
			return err
		}

		defer file.Close()		
		fmt.Printf("Reading all data...\n")
    	Data_r, err := ioutil.ReadAll(file)

		if(err != nil){
			fmt.Printf("File read error: %s\n", args.Sdfsname)
			(*reply) = 0
			return err
		}

		fmt.Printf("Calling RPC (%s:%s)...\n", args.Ip, args.Port)
		client, err := rpc.DialHTTP("tcp", args.Ip + ":" + args.Port) //Connect to given address
		if err != nil {
			fmt.Printf("Replicee %s is unreachable\n", args.Ip)
			return err
		}

		fmt.Printf("Replicating %s to address %s:%s\n", args.Sdfsname, args.Ip, args.Port)

		// Synchronous call***************************************
		var args = Write_args{Sdfsname: args.Sdfsname, Data: string(Data_r), Timestamp: int64(time.Now().Nanosecond())} //Create the args gob to send over to the RPC
		var reply int //Create a container for our results
		_ = client.Call("Sdfsrpc.Write_file", args, &reply) //Make the remote call
		return nil
	}
}
