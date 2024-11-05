package main

import (
	"fmt"
	_ "net/http/pprof"
	"time"
)

var RECORD_PROCESSOR_NUMBER int = 1
var NUMBER_OF_READERS = 8
var READ_BUFFER_SIZE = 8 * 1024 * 1024

// var TEST_FILE = "temperatures.csv"
//var TEST_FILE = "measurements.txt"

var TEST_FILE = "measurements_10M.txt"

// var TEST_FILE = "measurements_100M.txt"

// var TEST_FILE = "measurements_1M.txt"

func main() {
	// var wg sync.WaitGroup
	// go func() {
	// 	fmt.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	// wg.Add(1) // pprof - so we won't exit prematurely
	// wg.Add(1) // for the hardWork

	//slog.SetLogLoggerLevel(slog.LevelDebug.Level())

	//fmt.Printf("Going to read %v records with %v record processors\n", RECORD_PROCESSOR_NUMBER)
	// time.Sleep(5 * time.Second)
	startTime := time.Now()

	processFile(TEST_FILE)
	//wg.Wait()
	endTime := time.Now()
	fmt.Printf("Execution time is %v\n", endTime.Sub(startTime))
}

func processFile(s string) {

	splitFileReader := SplitFileReaders{nbReaders: NUMBER_OF_READERS, fileName: s}
	mapsChan, _ := splitFileReader.processFileConcurrently()

	for i := 0; i < NUMBER_OF_READERS; i++ {
		c := <-mapsChan
		fmt.Println(c)
	}
}
