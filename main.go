package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var NUMBER_OF_READERS = 8
var READ_BUFFER_SIZE = 8 * 1024 * 1024

// var TEST_FILE = "temperatures.csv"
var TEST_FILE = "measurements.txt"

// var TEST_FILE = "measurements_10M.txt"

// var TEST_FILE = "measurements_100M.txt"

// var TEST_FILE = "measurements_1M.txt"

func main() {
	//var wg sync.WaitGroup
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	//slog.SetLogLoggerLevel(slog.LevelDebug.Level())
	startTime := time.Now()
	processFile(TEST_FILE)
	endTime := time.Now()
	fmt.Printf("Execution time is %v\n", endTime.Sub(startTime))
}

func processFile(s string) {

	splitFileReader := SplitFileReaders{nbReaders: NUMBER_OF_READERS, fileName: s}
	reducer := NewReducer()
	mapsChan, _ := splitFileReader.processFileConcurrently()

	for i := 0; i < NUMBER_OF_READERS; i++ {
		partialResultMap := <-mapsChan
		reducer.reduce(partialResultMap)
	}

	fmt.Println(reducer.resultMap)
}
