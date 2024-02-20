package common

import (
	"fmt"
	"sync"
	"time"
)

const (
	EventCreate EventType = iota
	EventRemove
	EventUpdate
)

type EventType int

func (et EventType) String() string {
	switch et {
	case EventCreate:
		return "Create"
	case EventRemove:
		return "Remove"
	case EventUpdate:
		return "Update"
	default:
		panic(fmt.Sprintf("invalid event type: %d", et))
	}
}

type Event struct {
	Type  EventType
	Entry *Entry
}

func (ev Event) String() string {
	return fmt.Sprintf("%s: %s", ev.Type.String(), ev.Entry.String())
}

type Watcher struct {
	mutex   sync.Mutex
	files   []*Entry
	fileMap map[string]*Entry
	baseDir string
	options ListOptions
	events  chan []Event
}

func Watch(baseDir string, options ListOptions) *Watcher {
	out := &Watcher{
		mutex:   sync.Mutex{},
		files:   nil,
		baseDir: baseDir,
		options: options,
	}
	out.updateList()
	return out
}

func (w *Watcher) AddFilter(filter func(*Entry) bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.options.Filter == nil {
		w.options.Filter = filter
	} else {
		curFilter := w.options.Filter
		w.options.Filter = func(info *Entry) bool {
			return curFilter(info) && filter(info)
		}
	}

	var filteredFiles []*Entry
	for _, it := range w.files {
		if w.options.Filter(it) {
			filteredFiles = append(filteredFiles, it)
		}
	}
	w.files = filteredFiles
}

func (w *Watcher) Start(pollingInterval time.Duration) chan []Event {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.events != nil {
		panic("Watcher.Start was already called")
	}

	w.events = make(chan []Event)
	go func() {
		for {
			<-time.After(100 * time.Millisecond)
			if events := w.doScan(true); len(events) > 0 {
				w.events <- events
			}
		}
	}()
	return w.events
}

func (w *Watcher) List() []*Entry {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.files
}

func (w *Watcher) Scan() (out []Event) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.events != nil {
		panic("Watcher.Scan: manual scanning should not be used with Watcher.Start()")
	}
	return w.doScan(false)
}

func (w *Watcher) doScan(lock bool) (out []Event) {
	if lock {
		w.mutex.Lock()
		defer w.mutex.Unlock()
	}

	oldList, oldMap := w.files, w.fileMap
	w.updateList()
	newList, newMap := w.files, w.fileMap

	for _, old := range oldList {
		if new, ok := newMap[old.FullPath()]; !ok || new.IsDir != old.IsDir {
			out = append(out, Event{Type: EventRemove, Entry: old})
		}
	}

	for _, new := range newList {
		if old, ok := oldMap[new.FullPath()]; !ok || new.IsDir != old.IsDir {
			out = append(out, Event{Type: EventCreate, Entry: new})
		} else if new.ModTime() != old.ModTime() {
			out = append(out, Event{Type: EventUpdate, Entry: new})
		}
	}

	return
}

func (w *Watcher) updateList() {
	w.files = List(w.baseDir, w.options)
	if len(w.files) > 0 {
		w.fileMap = make(map[string]*Entry)
		for _, it := range w.files {
			w.fileMap[it.FullPath()] = it
		}
	} else {
		w.fileMap = nil
	}
}
