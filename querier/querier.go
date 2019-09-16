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

//A gob outline for the args we need for grep
type Args struct {
	Data, Filepath string
}

type Querier int //an alias for the type we need to handle this RPC

//The standard grep call function; takes arguments, puts them after grep, then puts the output of the grep in the reply pointer
//Method of Querier (which we need to do because RPC)
// Args = the struct of arguments to pass to the Grep
// 			{ 
// 				Data = the arguments for the grep, with flags and everything
// 				File = the file we want to search
// 			}
// reply = where we'll store our resulting string
func (t *Querier) Grep(args Args, reply *string) error {

	//Just a check-in for a handled request
	fmt.Printf("Args{Data = %s, Filepath = %s\n",args.Data,args.Filepath)
	s := strings.Fields(args.Data) //Make a slice of arguments

	//If there are no args, append a space so we can still do an empty grep
	if(len(s) == 0){
		s = append(s, " ")
	}

	//Append the filepath at the end for the grep
	s = append(s, args.Filepath)
	//Print the args
	fmt.Printf("Splice: %v\n\n", s)
	//Prepare grep with our slice of args
	cmd := exec.Command("grep", s...)//, args.Filepath)
	//Assign a fake stdout for us to collect results in
	var out bytes.Buffer
	cmd.Stdout = &out
	//Run the grep
	cmd.Run()

	//	Fuck errors we goin raw
	//   else if e != nil {
	// 		log.Fatal(e)
	// 		return e
	//   }

  	//Get the hostname so we can distinguish this machine
  	var name, e = os.Hostname()
  	if e != nil{
  		name = "irresolvable"
  	}

  	//Get the number of lines as per the spec
  	numlines := strings.Count(out.String(), "\n")
	str1 := strconv.Itoa(numlines)

	//Form our return string, and then return
    *reply = name + ":\n" + out.String() + "[" + str1 + "]\n";

	return nil
}