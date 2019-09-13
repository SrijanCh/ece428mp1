package test

import (
    "fmt"
    "io/ioutil"
    "bufio"
    "butes"
    "os"
    "container/list"
)

var ips list.list

file, err := os.Open("iptables.txt")
defer file.close()

if err != nil {
    fmt.Println("Error opening file")
    return 3
}

reader := bufio.NewReader(file)

var line string
for {
    line, err = reader.ReadString('\n')
    
    ips.PushBack(line)

    if err != nil {
        break
    }
