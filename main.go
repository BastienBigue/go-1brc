package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var NUMBER_OF_READERS = 8
var READ_BUFFER_SIZE = 8 * 1024 * 1024

// var TEST_FILE = IntputFile{"temperatures.csv", 10}

// var TEST_FILE = IntputFile{"draft.txt", 10}
// var TEST_FILE = IntputFile{"measurements_100.txt", 100}
// var TEST_FILE = IntputFile{"measurements_10k.txt", 10 * 1000}
// var TEST_FILE = IntputFile{"measurements_1M.txt", 1000 * 1000}
// var TEST_FILE = IntputFile{"measurements_10M.txt", 10 * 1000 * 1000}
// var TEST_FILE = InputFile{"measurements_100M.txt", 100* 1000 * 1000}
var TEST_FILE = IntputFile{"measurements.txt", 1000 * 1000 * 1000}

type IntputFile struct {
	fileName string
	nbLines  int64
}

func main() {

	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// slog.SetLogLoggerLevel(slog.LevelDebug.Level())

	startTime := time.Now()
	resultMap := processFile(TEST_FILE)
	endTime := time.Now()
	fmt.Printf("Execution time is %v\n", endTime.Sub(startTime))
	checkResult(TEST_FILE, resultMap)
}

func processFile(inputFile IntputFile) map[string]*MinMaxAverage {

	splitFileReader := NewSplitFileReader(inputFile.fileName, NUMBER_OF_READERS)
	reducer := NewReducer()
	mapsChan, _ := splitFileReader.processFileConcurrently()

	for i := 0; i < NUMBER_OF_READERS; i++ {
		partialResultMap := <-mapsChan
		reducer.reduce(partialResultMap)
	}

	fmt.Println(reducer.resultMap)
	return reducer.resultMap
}

func checkResult(inputFile IntputFile, resultMap map[string]*MinMaxAverage) {
	var nbLinesProcessed int64
	for _, v := range resultMap {
		nbLinesProcessed += int64(v.count)
	}
	if inputFile.nbLines != nbLinesProcessed {
		panic(fmt.Sprintf("Number of lines processed is incorrect ! Expected: %v, Got: %v", inputFile.nbLines, nbLinesProcessed))
	}
}
