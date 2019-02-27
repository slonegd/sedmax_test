package main

import "bufio"
import "fmt"
import "net"
import "net/http"
import "strings"
import "sync"
import "time"
import "flag"

// Value ...
type Value struct {
	value   string
	elapsed time.Time
}

var database = make(map[string]Value)
var dbMtx sync.Mutex

// Task ...
type Task struct {
	command string
	conn    net.Conn
}

func worker(tasks <-chan Task) {
	for t := range tasks {
		parseAndAnswer(t.command, t.conn)
	}
}

var tasks chan Task

func main() {
	taskQty := *flag.Int("task", 100, "max task for database changes")
	workersQty := *flag.Int("worker", 4, "max workers for database changes")
	HTTPport := *flag.Int("http", 8080, "port for http")
	TCPport := *flag.Int("tcp", 5000, "port for tcp")
	// livingTime := *flag.Int("time", 5, "living time for keys in second")
	flag.Parse()
	tasks = make(chan Task, taskQty)
	for i := 0; i < workersQty; i++ {
		go worker(tasks)
	}
	go listenTCP(TCPport)
	http.HandleFunc("/", handlerHTTP)
	http.ListenAndServe(fmt.Sprintf(":%d", HTTPport), nil)
}

func handlerHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{\n"))
	dbMtx.Lock()
	for key, v := range database {
		fmt.Fprintf(w, "  \"%s\" : \"%s\"\n", key, v.value)
	}
	dbMtx.Unlock()
	w.Write([]byte("}\n"))
}

func listenTCP(port int) {
	listener, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))

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
		tasks <- Task{str, conn}
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
		database[strs[key]] = Value{
			strs[value],
			time.Now().Add(time.Duration(1000)),
		}
		dbMtx.Unlock()
		write("OK", conn)
		return
	}
}

func write(s string, conn net.Conn) {
	_, errW := conn.Write([]byte(s + "\n"))
	if errW != nil {
		fmt.Println("cant write, close")
		conn.Close()
	}
}
