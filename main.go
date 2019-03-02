package main

import "bufio"
import "fmt"
import "net"
import "net/http"
import "flag"

var database *Database

func main() {
	taskQty := flag.Int("task", 100, "max task for database changes")
	workersQty := flag.Int("worker", 4, "max workers for database changes")
	HTTPport := flag.Int("http", 8080, "port for http")
	TCPport := flag.Int("tcp", 5000, "port for tcp")
	livingTime := flag.Int("time", 60, "living time for keys in second")
	flag.Parse()

	database = MakeDatabase(*taskQty, *workersQty, *livingTime)

	go listenTCP(*TCPport)
	http.HandleFunc("/", handlerHTTP)
	http.ListenAndServe(fmt.Sprintf(":%d", *HTTPport), nil)
}

func handlerHTTP(w http.ResponseWriter, r *http.Request) {
	database.HTTPresponse(w)
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
		database.AddTask(str, conn)
		fmt.Printf("%s\n", str)

	}
}

func write(s string, conn net.Conn) {
	_, errW := conn.Write([]byte(s + "\n"))
	if errW != nil {
		fmt.Println("cant write, close")
		conn.Close()
	}
}
