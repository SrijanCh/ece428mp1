package memtable
import(
	"fmt"
	"sync"
	"detector"
	"sort"
)

type Memtable struct{
	mu sync.Mutex
	table map[int]detector.Node_id_t
	hash_list []int
}


func NewMemtable() Memtable{
	var a sync.Mutex 
	b := make(map[int]detector.Node_id_t)
	c := make([]int, 0)
	return Memtable{a, b, c}
}

func (t *Memtable) Add_node(node_hash int, node_id detector.Node_id_t) {
	//Grab lock 
	t.mu.Lock()
	//Check if key exists
	if _, ok := t.table[node_hash]; !ok { //Node is not in there
		//Add to map
		t.table[node_hash] = node_id;
		//Append key to slice
		t.hash_list = append(t.hash_list, node_hash)
		//Sort the slice
		sort.Ints(t.hash_list);
	}

	//Release lock
	t.mu.Unlock()
}

func (t *Memtable) Delete_node(node_hash int, node_id detector.Node_id_t) {
	//Grab lock 
	t.mu.Lock()

	if _, ok := t.table[node_hash]; ok { //Node is in there
		//Delete from map
		delete(t.table, node_hash)
		//Find key in list, delete key by switching with last element, then re-sort
		i := 0
		for ; i < len(t.hash_list) && t.hash_list[i] != node_hash; i++ {
			// i++
		} 
		if i != len(t.hash_list){
    		t.hash_list[len(t.hash_list)-1], t.hash_list[i] = t.hash_list[i], t.hash_list[len(t.hash_list)-1]
    		t.hash_list = t.hash_list[:len(t.hash_list)-1]
			sort.Ints(t.hash_list);
		}
	}

	//Release lock
	t.mu.Unlock()
}


func (t *Memtable) Get_node(node_hash int) detector.Node_id_t{
	//Grab lock 
	// var a detector.Node_id_t = nil
	t.mu.Lock()
	//Get
	a, ok := t.table[node_hash]
	//Release lock
	t.mu.Unlock()
	
	if(ok){
		return a
	}else{
		var b []byte
		return detector.Node_id_t{0, b}
	}
}

//[pred1, pred2, succ1, succ2]
//leverages sorted list and >5 thing
func (t *Memtable) Get_neighbors(node_hash int) [4]int{
	
	ret := [4]int{-1,-1,-1,-1}
	ret[0] = -1
	ret[1] = -1
	ret[2] = -1
	ret[3] = -1
	//Grab lock

	// fmt.Printf("Get_neighbors\n")

	t.mu.Lock()

	if(len(t.hash_list) < 5){ //Not supposed to be functional cluster without 5 nodes
		// fmt.Printf("Fuck this shit, %d, %d, %d, %d\n", ret[0], ret[1], ret[2], ret[3] )
		t.mu.Unlock()
		return ret
	}

	i := 0
	for ; i < len(t.hash_list) && t.hash_list[i] != node_hash; i++ {
		// i++
	} 

	if(i >= len(t.hash_list)){ //Nonexistent
		//Do jack shit
	}else if(i == len(t.hash_list)-1){ //Wraparound (last)
		ret[0] = t.hash_list[i-2]
		ret[1] = t.hash_list[i-1]
		ret[2] = t.hash_list[0]
		ret[3] = t.hash_list[1]
	}else if(i == len(t.hash_list)-2){ //Wraparound (second to last)
		ret[0] = t.hash_list[i-2]
		ret[1] = t.hash_list[i-1]
		ret[2] = t.hash_list[i+1]
		ret[3] = t.hash_list[0]
	}else if(i == 0){ //Wraparound (first)
		ret[0] = t.hash_list[len(t.hash_list)-1]
		ret[1] = t.hash_list[len(t.hash_list)]
		ret[2] = t.hash_list[i+1]
		ret[3] = t.hash_list[i+2]
	}else if(i == 1){ //Wraparound {second}
		ret[0] = t.hash_list[len(t.hash_list)]
		ret[1] = t.hash_list[i-1]
		ret[2] = t.hash_list[i+1]
		ret[3] = t.hash_list[i+2]
	}else{	//Regular case
		ret[0] = t.hash_list[i-2]
		ret[1] = t.hash_list[i-1]
		ret[2] = t.hash_list[i+1]
		ret[3] = t.hash_list[i+2]
	}
	//Release lock
	t.mu.Unlock()
	return ret
}


func (t *Memtable) Get_num_nodes() int{
	a := -1
	t.mu.Lock()
	a = len(t.table)
	t.mu.Unlock()
	return a
}

func (t* Memtable) Get_avail_hash() int{
	t.mu.Lock()
	i := 0
	for ; i < len(t.hash_list) && t.hash_list[i] == i; i++ {
			fmt.Printf("[Get_avail_hash] i is %d; len(hash_list) is %d; the value at i is %d.", i, len(t.hash_list), t.hash_list[i])
	} 
			fmt.Printf("[Get_avail_hash] Broke with i %d.", i)
	t.mu.Unlock()

	// if(i == len(t.hash_list)){
	// 	return -1
	// }else{
		return i
	// }

}


func (t *Memtable) String() string{
	//Grab lock 
	// var a detector.Node_id_t = nil
	var ret string = ""
	t.mu.Lock()
	//Get
	ret += "Current Membership Table:\n"
	temp := ""
	for k, _ := range t.table{
		temp = fmt.Sprintf("Hash: %d, Node_Id: %s@%d\n", k, t.table[k].IPV4_addr.String(), t.table[k].Timestamp )
		ret += temp
	}
	ret += "Sorted Key List: "
	ret += fmt.Sprintf("%v\n", t.hash_list )
	//Release lock
	t.mu.Unlock()
	return ret
}