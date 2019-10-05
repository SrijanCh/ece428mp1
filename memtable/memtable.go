package memtable
import(
	"sync"
	"detector"
	"sort"
)

type Memtable struct{
	// mu_map sync.Mutex
	// mu_list sync.Mutex
	mu sync.Mutex
	table map[int]detector.Node_id_t
	hash_list []int
}

// type BeatTable struct{
// 	list map
// }


func (t *Memtable) Add_node(node_hash int, node_id detector.Node_id_t) {
	//Grab lock 
	t.mu.Lock()
	//Add to map
	t.table[node_hash] = node_id;
	//Append key to slice
	t.hash_list = append(t.hash_list, node_hash)
	//Sort the slice
	sort.Ints(t.hash_list);
	//Release lock
	t.mu.Unlock()
}

func (t *Memtable) Delete_node(node_hash int, node_id detector.Node_id_t) {
	//Grab lock 
	t.mu.Lock()

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

	//Release lock
	t.mu.Unlock()
}


func (t *Memtable) Get_node_id(node_hash int) detector.Node_id_t{
	//Grab lock 
	// var a detector.Node_id_t = nil
	t.mu.Lock()
	//Get
	a := t.table[node_hash]
	//Release lock
	t.mu.Unlock()
	return a
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

	t.mu.Lock()

	if(len(t.hash_list) < 5){ //Not supposed to be functional cluster without 5 nodes
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