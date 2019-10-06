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
}



func NewBeatTable() Beat_table{
	var a sync.Mutex 
	b := make(map[int]int64)
	return Beat_table{a, b}
}


func (t *Beat_table) Check_num() int{
	a := -1
	t.mu.Lock()
	a = len(t.table)
	t.mu.Unlock()
	return a
}

func (t *Beat_table) Log_beat(node_hash int, timestamp int64){
	t.mu.Lock()

	if _, ok := t.table[node_hash]; ok { //Node is in there
		t.table[node_hash] = timestamp
	}else{
		fmt.Printf("Just tried to log heartbeat someone not in table")
	}

	t.mu.Unlock()
}


func (t *Beat_table) Add_entry(node_hash int){
	t.mu.Lock()

	if _, ok := t.table[node_hash]; !ok { //Node is not there
		t.table[node_hash] = 0
	}else{
		fmt.Printf("Node is already in table!")
	}

	t.mu.Unlock()
}

func (t *Beat_table) Get_beat(node_hash int) int64{
	var a int64 = -1
	t.mu.Lock()
	if _, ok := t.table[node_hash]; ok { //Node is in there
		a = t.table[node_hash]
	}else{
		fmt.Printf("Just tried to get heartbeat of someone not in table")
	}
	t.mu.Unlock()
	return a
}

//Builds a table with new neighbors; 0 is timestamp assigned if not carried over, 
func (t *Beat_table) Reval_table(node_hash int, mem_table memtable.Memtable) [4]int{
	var neighbors [4]int = mem_table.Get_neighbors(node_hash)
	var newtable map[int]int64
	if newtable[0] == -1 || newtable[1] == -1 || newtable[2] == -1 || newtable[3] == -1{
		fmt.Printf("Can't get neighbors\n")
		return neighbors
	}

	t.mu.Lock()
	for i := 0; i < 4; i++{
		if _, ok := t.table[neighbors[i]]; ok { //Node is in there
			newtable[neighbors[i]] = t.table[neighbors[i]]
		}else{
			newtable[neighbors[i]] = 0
		}
	}
	t.table = newtable
	t.mu.Unlock()

	return neighbors
}
