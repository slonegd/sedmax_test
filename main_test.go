package main

import "testing"
import "net"
import "net/http/httptest"
import "bufio"
import "strings"
import "time"
import "io/ioutil"

func TCPcommand(conn net.Conn, c string, dontWaitAnswer ...bool) (out string) {
	in := bufio.NewReader(conn)
	conn.Write([]byte(c))
	if len(dontWaitAnswer) == 0 {
		out, _ = in.ReadString('\n')
	} else {
		out = ""
	}
	return
}
func TestInsertNew(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	answer := TCPcommand(client, "INSERT 1 1\r\n")
	if strings.Compare(answer, "OK\n") != 0 {
		t.Errorf("Expected OK got %s", answer)
	}

	answer = TCPcommand(client, "INSERT 1 2\n")
	if strings.Compare(answer, "OK\n") != 0 {
		t.Errorf("Expected OK got %s", answer)
	}
	client.Close()
}

func TestInsertDouble(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	answer := TCPcommand(client, "INSERT 1 1\n")
	if strings.Compare(answer, "OK\n") != 0 {
		t.Errorf("Expected OK got %s", answer)
	}

	dontWaitAnswer := true
	answer = TCPcommand(client, "INSERT 1 1\n", dontWaitAnswer)
	if strings.Compare(answer, "") != 0 {
		t.Errorf("Expected empty got %s", answer)
	}
	client.Close()
}

func TestGetExist(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	TCPcommand(client, "INSERT 1 2\n")
	answer := TCPcommand(client, "GET 1\n")
	if strings.Compare(answer, "2\n") != 0 {
		t.Errorf("Expected 2 got %s", answer)
	}
	client.Close()
}

func TestGetDontExist(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	TCPcommand(client, "INSERT 1 1\n")
	answer := TCPcommand(client, "GET 2\n")
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestGetElapsed(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	TCPcommand(client, "INSERT 1 2\n")
	time.Sleep(1 * time.Second)
	answer := TCPcommand(client, "GET 1\n")
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestGetUpdated(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	TCPcommand(client, "INSERT 1 2\n")
	time.Sleep(500 * time.Millisecond)
	TCPcommand(client, "INSERT 1 3\n")
	time.Sleep(750 * time.Millisecond)
	answer := TCPcommand(client, "GET 1\n")
	if strings.Compare(answer, "3\n") != 0 {
		t.Errorf("Expected 3 got %s", answer)
	}
	time.Sleep(250 * time.Millisecond)
	answer = TCPcommand(client, "GET 1\n")
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestGetNotUpdated(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	TCPcommand(client, "INSERT 1 2\n")
	time.Sleep(500 * time.Millisecond)
	TCPcommand(client, "INSERT 1 2\n", true)
	time.Sleep(500 * time.Millisecond)
	client.Write([]byte("GET 1\n"))
	answer := TCPcommand(client, "GET 1\n")
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestDeleteExist(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	TCPcommand(client, "INSERT 1 2\n")
	answer := TCPcommand(client, "GET 1\n")
	if strings.Compare(answer, "2\n") != 0 {
		t.Errorf("Expected 2 got %s", answer)
	}
	answer = TCPcommand(client, "DELETE 1\n")
	if strings.Compare(answer, "OK\n") != 0 {
		t.Errorf("Expected OK got %s", answer)
	}
	answer = TCPcommand(client, "GET 1\n")
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestDeleteDontExist(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	answer := TCPcommand(client, "DELETE 1\n")
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestDeleteElapsed(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	TCPcommand(client, "INSERT 1 2\n")
	time.Sleep(1 * time.Second)
	answer := TCPcommand(client, "DELETE 1\n")
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestCloseTCP(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()

	client.Write([]byte("INSERT 1 2\n"))
	client.Close()
}

func emptyRequest(database *Database) string {
	request := httptest.NewRequest("", "/", nil)
	recoder := httptest.NewRecorder()
	database.HTTPresponse(recoder, request)
	response := recoder.Result()
	body, _ := ioutil.ReadAll(response.Body)
	return string(body)
}

func TestHTTPempty(t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)

	expected := `{
}
`
	got := emptyRequest(database)
	if strings.Compare(expected, got) != 0 {
		t.Errorf("Expected %s got %s", expected, got)
	}
}

func TestHTTP(t *testing.T) {
	maxTasks := 10
	workerQty := 1
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()
	defer client.Close()

	TCPcommand(client, "INSERT 1 2\n")
	TCPcommand(client, "INSERT 3 4\n")

	expected1 := `{
  "1" : "2",
  "3" : "4"
}
`
	expected2 := `{
  "3" : "4",
  "1" : "2"
}
`
	got := emptyRequest(database)
	if !(strings.Compare(expected1, got) == 0 || strings.Compare(expected2, got) == 0) {
		t.Errorf("Expected %s or %s got %s", expected1, expected2, got)
	}
}

func TestHTTPdeleted(t *testing.T) {
	maxTasks := 10
	workerQty := 1
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()
	defer client.Close()

	TCPcommand(client, "INSERT 1 2\n")
	TCPcommand(client, "INSERT 3 4\n")

	expected1 := `{
  "1" : "2",
  "3" : "4"
}
`
	expected2 := `{
  "3" : "4",
  "1" : "2"
}
`
	got := emptyRequest(database)
	if !(strings.Compare(expected1, got) == 0 || strings.Compare(expected2, got) == 0) {
		t.Errorf("Expected %s or %s got %s", expected1, expected2, got)
	}

	TCPcommand(client, "DELETE 3\n")

	expected := `{
  "1" : "2"
}
`
	got = emptyRequest(database)
	if strings.Compare(expected, got) != 0 {
		t.Errorf("Expected %s got %s", expected, got)
	}
}

func TestHTTPelapsed(t *testing.T) {
	maxTasks := 10
	workerQty := 1
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server, database)
		server.Close()
	}()
	defer client.Close()

	TCPcommand(client, "INSERT 1 2\n")
	time.Sleep(500 * time.Millisecond)
	TCPcommand(client, "INSERT 3 4\n")

	expected1 := `{
  "1" : "2",
  "3" : "4"
}
`
	expected2 := `{
  "3" : "4",
  "1" : "2"
}
`
	got := emptyRequest(database)
	if !(strings.Compare(expected1, got) == 0 || strings.Compare(expected2, got) == 0) {
		t.Errorf("Expected %s or %s got %s", expected1, expected2, got)
	}

	time.Sleep(500 * time.Millisecond)

	expected := `{
  "3" : "4"
}
`
	got = emptyRequest(database)
	if strings.Compare(expected, got) != 0 {
		t.Errorf("Expected %s got %s", expected, got)
	}
}
