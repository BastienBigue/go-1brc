package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"
)

var NUMBER_OF_READERS = 8
var READ_BUFFER_SIZE = 8 * 1024 * 1024

// var TEST_FILE = IntputFile{"measurements_10.txt", 10}

// var TEST_FILE = IntputFile{"measurements_1k.txt", 1000}
// var TEST_FILE = IntputFile{"measurements_100k.txt", 100 * 1000}
// var TEST_FILE = IntputFile{"measurements_10M.txt", 10 * 1000 * 1000}
var TEST_FILE = IntputFile{"measurements_1B.txt", 1000 * 1000 * 1000}

type IntputFile struct {
	fileName string
	nbLines  int64
}

func main() {

	f, err := os.Create("cpu_profile.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

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
