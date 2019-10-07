package beat_table
import(
	"fmt"
	"sync"
	// "detector"
	"memtable"
	// "sort"
)

type Beat_table struct{
	mu sync.Mutex
	table map[int]int64
	count int64
}



func NewBeatTable() Beat_table{
	fmt.Printf("NewBeatTable----------------------------------------\n")
	var a sync.Mutex 
	b := make(map[int]int64)
	c := 1
	return Beat_table{a, b, int64(c)}
}


func (t *Beat_table) Check_num() int{
	fmt.Printf("Check_num----------------------------------------\n")
	a := -1
	t.mu.Lock()
	a = len(t.table)
	t.mu.Unlock()
	fmt.Printf(t.String())
	return a
}


func (t *Beat_table) String() string{
	t.mu.Lock()
	ret := "Current Beat Table:\n"
	temp := ""
	for k, _ := range t.table{
		temp = fmt.Sprintf("Hash: %d, Timestamp: %d\n", k, t.table[k])
		ret += temp
	}
	t.mu.Unlock()
	return ret
}

func (t *Beat_table) Log_beat(node_hash int, timestamp int64){
	fmt.Printf("Log_beat----------------------------------------\n")
	if timestamp == 0{
		fmt.Printf("===========================================LOGGING A TIMESTAMP OF ZERO========================================================\n")
	}

	t.mu.Lock()

	if _, ok := t.table[node_hash]; ok { //Node is in there
		fmt.Printf("Set timestamp at %d in table from %d to %d\n", node_hash, t.table[node_hash], timestamp)
		t.table[node_hash] = timestamp
	}else{
		fmt.Printf("Just tried to log heartbeat someone not in table (%d)\n", node_hash)
	}

	t.mu.Unlock()

	fmt.Printf(t.String())
}


func (t *Beat_table) Add_entry(node_hash int){
	fmt.Printf("Add_entry----------------------------------------\n")
	t.mu.Lock()

	if _, ok := t.table[node_hash]; !ok { //Node is not there
		t.table[node_hash] = 0
	}else{
		fmt.Printf("Node is already in table!")
	}

	t.mu.Unlock()
}

func (t *Beat_table) Get_beat(node_hash int) int64{
	fmt.Printf("Get_beat----------------------------------------\n")
	var a int64 = -1
	t.mu.Lock()
	if _, ok := t.table[node_hash]; ok { //Node is in there
		a = t.table[node_hash]
	}else{
		fmt.Printf("Just tried to get heartbeat of someone not in table (%d)\n", node_hash)
	}
	t.mu.Unlock()
	if a == 0{
		fmt.Printf("[Get_beat with %d]======================================RETURNING A TIMESTAMP OF ZERO=================================================\n", t.table[node_hash])
	}

	fmt.Printf(t.String())
	return a
}

//Builds a table with new neighbors; 0 is timestamp assigned if not carried over, 
func (t *Beat_table) Reval_table(node_hash int, mem_table memtable.Memtable) [4]int{
	// fmt.Printf("Reval_table----------------------------------------\n")
	var neighbors [4]int = mem_table.Get_neighbors(node_hash)
	
	if neighbors[0] == -1 || neighbors[1] == -1 || neighbors[2] == -1 || neighbors[3] == -1{
		// fmt.Printf("Can't get neighbors\n")
		return neighbors
	}

	fmt.Printf("1. Making new temporary map\n")
	var newtable map[int]int64 = make(map[int]int64)
	for k, v := range newtable{
		fmt.Printf("Value %d, %d\n", k, v)
	}

	t.mu.Lock()
	for i := 0; i < 4; i++{
		fmt.Printf("2. Looping through neighbors and checking/transferring values; neighbors: %v, i: %d", neighbors, i)
		for k, v := range newtable{
			fmt.Printf("Value %d, %d\n", k, v)
		}
		fmt.Printf(t.String())

		if _, ok := t.table[neighbors[i]]; ok { //Node is in there
			fmt.Printf("3. Node at %d already in here, transfer value %d", neighbors[i], t.table[i])
			fmt.Printf("Before:\n")
			for k, v := range newtable{
				fmt.Printf("Value %d, %d\n", k, v)
			}
			fmt.Printf(t.String())

			newtable[neighbors[i]] = t.table[i]
			
			fmt.Printf("After:\n")
			for k, v := range newtable{
				fmt.Printf("Value %d, %d\n", k, v)
			}
			fmt.Printf(t.String())
		}else{
			if t.count == 0{
				t.count = 1
			}
			fmt.Printf("3.============================NEW NEIGHBOR %d STARTED WITH COUNT %d=========================", neighbors[i], t.count)
			fmt.Printf("Before:\n")
			for k, v := range newtable{
				fmt.Printf("Value %d, %d\n", k, v)
			}
			fmt.Printf(t.String())

			newtable[neighbors[i]] = t.count

			fmt.Printf("After:\n")
			for k, v := range newtable{
				fmt.Printf("Value %d, %d\n", k, v)
			}
			fmt.Printf(t.String())
		}
	}
	// t.table = newtable

	//Clear the map
			fmt.Printf("4. Clear our old map\n")
			fmt.Printf("Before:\n")
			for k, v := range newtable{
				fmt.Printf("Value %d, %d\n", k, v)
			}
			fmt.Printf(t.String())
	for k,_ := range t.table{
		delete(t.table, k)
	}
			fmt.Printf("After:\n")
			for k, v := range newtable{
				fmt.Printf("Value %d, %d\n", k, v)
			}
			fmt.Printf(t.String())

	//Copy in the new table

			fmt.Printf("5. Copy into our map\n")
			fmt.Printf("Before:\n")
			for k, v := range newtable{
				fmt.Printf("Value %d, %d\n", k, v)
			}
			fmt.Printf(t.String())
	for k,v := range newtable{
			fmt.Printf("Copying %d to %d\n", v, k)
		t.table[k] = v
	}
			fmt.Printf("After:\n")
			for k, v := range newtable{
				fmt.Printf("Value %d, %d\n", k, v)
			}
			fmt.Printf(t.String())

	t.count = (t.count+1) % 3000
	t.mu.Unlock()

			fmt.Printf("Result:\n")
	fmt.Printf(t.String())

	return neighbors
}
