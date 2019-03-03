package main

import "bufio"
import "fmt"
import "net"
import "net/http"
import "flag"

func main() {
	taskQty := flag.Int("task", 100, "max task for database changes")
	workersQty := flag.Int("worker", 4, "max workers for database changes")
	HTTPport := flag.Int("http", 8080, "port for http")
	TCPport := flag.Int("tcp", 5000, "port for tcp")
	livingTime := flag.Int("time", 60, "living time for keys in second")
	flag.Parse()

	database := MakeDatabase(*taskQty, *workersQty, *livingTime)

	go listenTCP(*TCPport, database)
	http.HandleFunc("/", database.HTTPresponse)
	http.ListenAndServe(fmt.Sprintf(":%d", *HTTPport), nil)
}

func listenTCP(port int, database *Database) {
	listener, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))

	for {
		conn, err := listener.Accept()
		if err != nil {
			conn.Close()
			continue
		}
		go handlerTCP(conn, database)
	}
}

func handlerTCP(conn net.Conn, database *Database) {
	defer conn.Close()
	for {
		in := bufio.NewReader(conn)
		str, errR := in.ReadString('\n')
		if errR != nil {
			break
		}
		str = str[:len(str)-1]
		if str[len(str)-1] == '\r' {
			str = str[:len(str)-1]
		}
		database.AddTask(str, conn)
	}
}

func write(s string, conn net.Conn) {
	_, errW := conn.Write([]byte(s + "\n"))
	if errW != nil {
		conn.Close()
	}
}
