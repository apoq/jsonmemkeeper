package jsonmemkeep

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"os"
	"sync"
)

type Keeper struct {
	fileName     string
	outputStruct interface{}
	watcher      *fsnotify.Watcher
	mu           *sync.RWMutex
}

func NewListener(jsonFileName string, outputStruct interface{}) *Keeper {
	watcher, _ := fsnotify.NewWatcher()
	watcher.Add(jsonFileName)

	return &Keeper{
		fileName:     jsonFileName,
		outputStruct: outputStruct,
		watcher:      watcher,
		mu:           new(sync.RWMutex),
	}
}

func (k *Keeper) Close() {
	k.watcher.Close()
}

func (k *Keeper) Run() {
	k.mu.Lock()
	defer k.mu.Unlock()
	err := k.jsonFileToStruct(k.fileName, k.outputStruct)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case event, ok := <-k.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					k.mu.Lock()
					err = k.jsonFileToStruct(k.fileName, k.outputStruct)
					k.mu.Unlock()
					if err != nil {
						panic(err)
					}
				}
			case err, ok := <-k.watcher.Errors:
				if !ok {
					return
				}
				panic(err)
			}
		}
	}()
}

func (k *Keeper) jsonFileToStruct(jsonFileName string, outputStruct interface{}) error {
	jsonFile, err := os.Open(jsonFileName)
	defer jsonFile.Close()
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &outputStruct)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keeper) Fetch() interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.outputStruct
}
