package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
)

func main() {
	go listen()
	http.HandleFunc("/", httpHandler)
	http.ListenAndServe(":8080", nil)
}

var db = make(map[string]string)

func httpHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{\n"))
	for key, v := range db {
		fmt.Fprintf(w, "  \"%s\" : \"%s\"\n", key, v)
	}
	w.Write([]byte("}\n"))
}
func listen() {
	listener, _ := net.Listen("tcp", ":5000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("cant connect")
			conn.Close()
			continue
		}
		fmt.Println("connect")
		go server(conn)
	}
}
func server(conn net.Conn) {
	defer conn.Close()
	for {
		in := bufio.NewReader(conn)
		str, err := in.ReadString('\r')
		if err != nil {
			fmt.Println("close")
			break
		}
		str = str[:len(str)-1]
		go parser(str)
		fmt.Printf("%sn", str)
		conn.Write([]byte("OK " + str + "\n"))
	}
}


func parser(in string) {
	strs := strings.Split(in," ")
	const (
		command = iota
		key
		value
	)

	if strings.Compare(strs[command], "INSERT") == 0 {
		db[strs[key]] = strs[value]
		return
	}
}