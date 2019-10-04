package detector
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
	// Data, Filepath string
}

type Detector int; //an alias for the type we need to handle this RPC

//The standard grep call function; takes arguments, puts them after grep, then puts the output of the grep in the reply pointer
//Method of Querier (which we need to do because RPC)
// Args = the struct of arguments to pass to the Grep
// 			{ 
// 				Data = the arguments for the grep, with flags and everything
// 				File = the file we want to search
// 			}
// reply = where we'll store our resulting string
func (t *Detector) Add_to_List(args Args, reply *string) error {

	
	return nil
}