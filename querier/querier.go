package querier
import (
		// "fmt"
		"bytes"
    	"os/exec"
    	"os"
	    // "log"
	    "fmt"
	    "strings"
		"strconv"
		)

type Args struct {
	Data, Filepath string
}

type Querier int

func (t *Querier) Grep(args Args, reply *string) error {

	fmt.Printf("Args{Data = %s, Filepath = %s\n",args.Data,args.Filepath)
	s := strings.Fields(args.Data)
	// s = append(s, "")
	s = append(s, args.Filepath)
	fmt.Printf("Splice: %v\n\n", s)
	cmd := exec.Command("grep", s...)//, args.Filepath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
  //	Fuck errors we goin raw
  //   else if e != nil {
		// log.Fatal(e)
		// return e
  //   }
  	var name, e = os.Hostname()
  	if e != nil{
  		name = "irresolvable"
  	}
  	numlines := strings.Count(out.String(), "\n")
	str1 := strconv.Itoa(numlines)
    *reply = name + ":\n" + out.String() + "[" + str1 + "]\n";
	return nil
}