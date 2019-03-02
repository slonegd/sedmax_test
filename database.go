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
	livingTime time.Duration
	tasks  chan task
}

// MakeDatabase ...
func MakeDatabase(taskQty, workersQty, livingTime int) (p *Database) {
	p = &Database{
		values: make(map[string]Value),
		livingTime : time.Duration(livingTime*1000000000), // in seconds
		tasks:  make(chan task, taskQty),
	}
	for i := 0; i < workersQty; i++ {
		go worker(p.tasks, p)
	}
	return
}

type task struct {
	command string
	conn    net.Conn
}

func worker(tasks <-chan task, d *Database) {
	for t := range tasks {
		d.parseAndAnswer(t)
	}
}

// HTTPresponse ...
func (d *Database) HTTPresponse(w http.ResponseWriter) {
	now := time.Now()
	w.Write([]byte("{"))
	notFirst := false
	d.Lock()
	for key, v := range d.values {
		if v.elapsed.After(now) {
			if notFirst {
				fmt.Fprintf(w, ",")
			}
			fmt.Fprintf(w, "\n  \"%s\" : \"%s\"", key, v.value)
			notFirst = true
		} else {
			delete(d.values, key)
		}
	}
	d.Unlock()
	w.Write([]byte("\n}\n"))
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

	if strings.Compare(strs[command], "INSERT") == 0 && len(strs) == 3 {
		d.addValue(strs[key], strs[value], t.conn)
		return
	}

	if strings.Compare(strs[command], "GET") == 0 && len(strs) == 2 {
		d.getValue(strs[key], t.conn)
		return
	}

	if strings.Compare(strs[command], "DELETE") == 0 && len(strs) == 2 {
		d.deleteValue(strs[key], t.conn)
		return
	}
}

func (d *Database) addValue(key, value string, conn net.Conn) {
	d.Lock()
	defer d.Unlock()
	v, exist := d.values[key]
	switch {
	case !exist:
		d.values[key] = Value{
			value,
			time.Now().Add(d.livingTime),
		}
		write("OK", conn)
	case v.value != value:
		d.values[key] = Value{
			value,
			time.Now().Add(d.livingTime),
		}
		write("OK", conn)
	}
	return
}

func (d *Database) getValue(key string, conn net.Conn) {
	now := time.Now()
	d.Lock()
	defer d.Unlock()
	v, exist := d.values[key]
	if exist && v.elapsed.After(now) {
		write(v.value, conn)
	} else {
		write("ERR", conn)
	}
	return
}

func (d *Database) deleteValue(key string, conn net.Conn) {
	now := time.Now()
	d.Lock()
	defer d.Unlock()
	v, exist := d.values[key]
	if exist {
		if v.elapsed.After(now) {
			write("OK", conn)
		} else {
			write("ERR", conn)
		}
		delete(d.values, key)
	} else {
		write("ERR", conn)
	}
	return
}
