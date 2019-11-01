package members

// import(
	
// )

type MemberNode struct {
	Ip string
	Timestamp int64
	Alive bool
}

var MemberMap = make(map[int]*MemberNode)
var Garbage = make(map[int]bool)