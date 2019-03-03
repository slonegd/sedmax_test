package main

import "testing"
import "net"
import "bufio"
import "strings"
import "time"

func TestInsertNew (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	client.Write([]byte("INSERT 1 1\r\n"))
	in := bufio.NewReader(client)
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "OK\n") != 0 {
		t.Errorf("Expected OK got %s", answer)
	}
	client.Write([]byte("INSERT 2 2\n"))
	answer, _ = in.ReadString('\n')
	if strings.Compare(answer, "OK\n") != 0 {
		t.Errorf("Expected OK got %s", answer)
	}
	client.Close()
}

func TestInsertDouble (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	client.Write([]byte("INSERT 1 1\n"))
	in := bufio.NewReader(client)
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "OK\n") != 0 {
		t.Errorf("Expected OK got %s", answer)
	}

	client.Write([]byte("INSERT 2 2\n"))
	answer, _ = in.ReadString('\n')
	if strings.Compare(answer, "OK\n") != 0 {
		t.Errorf("Expected OK got %s", answer)
	}
	client.Close()
}

func TestGetExist (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	client.Write([]byte("INSERT 1 2\n"))
	in := bufio.NewReader(client)
	in.ReadString('\n')
	client.Write([]byte("GET 1\n"))
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "2\n") != 0 {
		t.Errorf("Expected 2 got %s", answer)
	}
	client.Close()
}

func TestGetDontExist (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	client.Write([]byte("INSERT 1 2\n"))
	in := bufio.NewReader(client)
	in.ReadString('\n')
	client.Write([]byte("GET 2\n"))
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestGetElapsed (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	client.Write([]byte("INSERT 1 2\n"))
	in := bufio.NewReader(client)
	in.ReadString('\n')
	time.Sleep(1*time.Second)
	client.Write([]byte("GET 1\n"))
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestGetUpdated (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	in := bufio.NewReader(client)
	client.Write([]byte("INSERT 1 2\n"))
	in.ReadString('\n')
	time.Sleep(500*time.Millisecond)
	client.Write([]byte("INSERT 1 3\n"))
	in.ReadString('\n')
	time.Sleep(750*time.Millisecond)
	client.Write([]byte("GET 1\n"))
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "3\n") != 0 {
		t.Errorf("Expected 3 got %s", answer)
	}
	time.Sleep(250*time.Millisecond)
	client.Write([]byte("GET 1\n"))
	answer, _ = in.ReadString('\n')
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestGetNotUpdated (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	in := bufio.NewReader(client)
	client.Write([]byte("INSERT 1 2\n"))
	in.ReadString('\n')
	time.Sleep(500*time.Millisecond)
	client.Write([]byte("INSERT 1 2\n"))
	// in.ReadString('\n') // no answer
	time.Sleep(500*time.Millisecond)
	client.Write([]byte("GET 1\n"))
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestDeleteExist (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	client.Write([]byte("INSERT 1 2\n"))
	in := bufio.NewReader(client)
	in.ReadString('\n')
	client.Write([]byte("GET 1\n"))
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "2\n") != 0 {
		t.Errorf("Expected 2 got %s", answer)
	}
	client.Write([]byte("DELETE 1\n"))
	answer, _ = in.ReadString('\n')
	if strings.Compare(answer, "OK\n") != 0 {
		t.Errorf("Expected OK got %s", answer)
	}
	client.Write([]byte("GET 1\n"))
	answer, _ = in.ReadString('\n')
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestDeleteDontExist (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	client.Write([]byte("DELETE 1\n"))
	in := bufio.NewReader(client)
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestDeleteElapsed (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	client.Write([]byte("INSERT 1 2\n"))
	in := bufio.NewReader(client)
	in.ReadString('\n')
	time.Sleep(1*time.Second)
	client.Write([]byte("DELETE 1\n"))
	answer, _ := in.ReadString('\n')
	if strings.Compare(answer, "ERR\n") != 0 {
		t.Errorf("Expected ERR got %s", answer)
	}
	client.Close()
}

func TestCloseTCP (t *testing.T) {
	maxTasks := 10
	workerQty := 2
	livingTime := 1
	database := MakeDatabase(maxTasks, workerQty, livingTime)
	server, client := net.Pipe()
	go func() {
		handlerTCP(server,database)
		server.Close()
	}()

	client.Write([]byte("INSERT 1 2\n"))
	client.Close()
}