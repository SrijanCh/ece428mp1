package querier
import (
		"fmt"
		"bytes"
    	"os/exec"
	    "log"
		)

type Args struct {
	Data, Filepath string
}

type Querier int

func (t *Querier) Grep(args Args, reply *string) error {
	fmt.Printf("Args{Data = %s, Filepath = %s\n",args.Data,args.Filepath)
	cmd := exec.Command("grep", args.Data, args.Filepath)
	var out bytes.Buffer
	cmd.Stdout = &out
	e := cmd.Run()
    if e != nil {
		log.Fatal(e)
		return e
    }
    *reply = out.String();
	return nil
}