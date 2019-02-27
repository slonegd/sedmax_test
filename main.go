package main

import "bufio"
import "fmt"
import "net"
import "net/http"
import "strings"
import "sync"

var database = make(map[string]string)
var dbMtx sync.Mutex

// Task ...
type Task struct {
	command string
	conn net.Conn
}

func worker(tasks <-chan Task) {
    for t := range tasks {
        parseAndAnswer(t.command, t.conn)
    }
}
var tasks chan Task

func main() {
	tasks = make(chan Task, 100)
	for i:=0; i<4; i++ {
		go worker(tasks)
	}
	go listenTCP()
	http.HandleFunc("/", handlerHTTP)
	http.ListenAndServe(":8080", nil)
}

func handlerHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{\n"))
	dbMtx.Lock()
	for key, v := range database {
		fmt.Fprintf(w, "  \"%s\" : \"%s\"\n", key, v)
	}
	dbMtx.Unlock()
	w.Write([]byte("}\n"))
}

func listenTCP() {
	listener, _ := net.Listen("tcp", ":5000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("cant connect")
			conn.Close()
			continue
		}
		fmt.Println("connect")
		go connHandler(conn)
	}
}

func connHandler(conn net.Conn) {
	defer conn.Close()
	for {
		in := bufio.NewReader(conn)
		str, errR := in.ReadString('\r')
		if errR != nil {
			fmt.Println("cant read, close")
			break
		}
		str = str[:len(str)-1]
		tasks <- Task{str,conn}
		fmt.Printf("%s\n", str)
		
	}
}

func parseAndAnswer(in string, conn net.Conn) {
	strs := strings.Split(in, " ")
	const (
		command = iota
		key
		value
	)

	if strings.Compare(strs[command], "INSERT") == 0 {
		dbMtx.Lock()
		database[strs[key]] = strs[value]
		dbMtx.Unlock()
		write("OK",conn)
		return
	}
}

func write(s string, conn net.Conn) {
	_, errW := conn.Write([]byte(s+"\n"))
	if errW != nil {
		fmt.Println("cant write, close")
		conn.Close()
	}
}
