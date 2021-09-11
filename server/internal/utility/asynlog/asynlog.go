// Package asynlog provides an asynchronous logger.
package asynlog

import (
	"container/list"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// Logger is the interface of the logger.
type Logger interface {

	// Log prints a message immediately.
	Log(msg interface{})

	// Fatal prints the current and all previous messages, then exits.
	Fatal(msg interface{})

	// Panic prints the current and all previous messages, then throws a panic.
	Panic(msg interface{})

	// Store stores a message.
	Store(msg interface{}) int

	// LogStorage prints all previous stored messages.
	LogStorage()
}

type logger struct {
	msgs       *list.List
	sys        *log.Logger
	timeLayout string
	mutex      sync.Mutex
}

// New creates a new logger.
func New(out io.Writer, timeLayout string) Logger {
	return &logger{
		msgs:       list.New(),
		sys:        log.New(out, "", 0),
		timeLayout: timeLayout,
	}
}

func (log *logger) Log(msg interface{}) {
	log.sys.Println(log.format(msg))
}

func (log *logger) Fatal(msg interface{}) {
	log.LogStorage()
	log.sys.Fatalln(log.format(msg))
}

func (log *logger) Panic(msg interface{}) {
	log.LogStorage()
	log.sys.Panicln(log.format(msg))
}

func (log *logger) Store(msg interface{}) int {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	log.msgs.PushBack(log.format(msg))
	return log.msgs.Len()
}

func (log *logger) LogStorage() {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	for e := log.msgs.Front(); e != nil; e = e.Next() {
		log.sys.Println(e.Value)
	}

	log.msgs = log.msgs.Init()
}

func (log *logger) format(msg interface{}) string {
	if log.timeLayout != "" {
		return fmt.Sprintf("%s: %v", time.Now().Format(log.timeLayout), msg)
	}

	return fmt.Sprintf("%v", msg)
}
