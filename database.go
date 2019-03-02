package main

import "time"
import "sync"
import "net"
import "net/http"
import "fmt"
import "strings"

// Value ...
type Value struct {
	value   string
	elapsed time.Time
}

// Database ...
type Database struct {
	sync.Mutex
	values map[string]Value
	tasks  chan task
}

// MakeDatabase ...
func MakeDatabase(taskQty, workersQty int) (p *Database) {
	p = &Database{
		values: make(map[string]Value),
		tasks:  make(chan task, taskQty),
	}
	for i := 0; i < workersQty; i++ {
		go worker(p.tasks, p)
	}
	return
}

// Task ...
type task struct {
	command string
	conn    net.Conn
}

func worker(tasks <-chan task, d *Database) {
	for t := range tasks {
		d.parseAndAnswer(t)
	}
}

// HTTP ...
func (d *Database) HTTP(w http.ResponseWriter) {
	w.Write([]byte("{\n"))
	d.Lock()
	for key, v := range d.values {
		fmt.Fprintf(w, "  \"%s\" : \"%s\"\n", key, v.value)
	}
	d.Unlock()
	w.Write([]byte("}\n"))
}

// AddTask ...
func (d *Database) AddTask(command string, conn net.Conn) {
	d.tasks <- task {command, conn}
}

func (d *Database) parseAndAnswer(t task) {
	strs := strings.Split(t.command, " ")
	const (
		command = iota
		key
		value
	)

	if strings.Compare(strs[command], "INSERT") == 0 {
		d.Lock()
		d.values[strs[key]] = Value{
			strs[value],
			time.Now().Add(time.Duration(1000)),
		}
		d.Unlock()
		write("OK", t.conn)
		return
	}
}
