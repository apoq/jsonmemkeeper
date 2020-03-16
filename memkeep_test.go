package jsonmemkeep

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

const TestFileName = "test.json"

type TestStruct struct {
	Count int `json:"count"`
}

func TestWatcher(t *testing.T) {
	resetFile(TestFileName)
	assrt := assert.New(t)

	listener := NewListener(TestFileName, new(TestStruct))
	defer listener.Close()
	listener.Run()

	testStruct := listener.Fetch().(*TestStruct)
	assrt.Equal(0, testStruct.Count)

	incrementCounter(TestFileName)
	time.Sleep(1 * time.Second)

	testStruct = listener.Fetch().(*TestStruct)
	assrt.Equal(1, testStruct.Count)

	// CONCURRENT READ-WRITE TEST
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 1000000; i++ {
			testStruct = listener.Fetch().(*TestStruct)
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		for i := 0; i < 1000; i++ {
			incrementCounter(TestFileName)
		}
		wg.Done()
	}()
	wg.Wait()
	assrt.NotEqual(1001, testStruct.Count)
}

func incrementCounter(fileName string) {
	var testStruct = new(TestStruct)
	jsonFile, _ := os.OpenFile(fileName, os.O_RDWR, 0644)
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(bytes, &testStruct)

	testStruct.Count++
	newJson, _ := json.Marshal(testStruct)
	result := string(newJson) + "              "
	jsonFile.WriteAt([]byte(result), 0)
}

func resetFile(fileName string) {
	var testStruct = new(TestStruct)
	jsonFile, _ := os.OpenFile(fileName, os.O_RDWR, 0644)
	defer jsonFile.Close()
	testStruct.Count = 0
	newJson, _ := json.Marshal(testStruct)
	result := string(newJson) + "              "
	jsonFile.WriteAt([]byte(result), 0)
}
