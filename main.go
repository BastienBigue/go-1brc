package main

import (
	"fmt"
	_ "net/http/pprof"
	"runtime"
	"slices"
	"time"
)

var NUMBER_OF_READERS = runtime.NumCPU()
var READ_BUFFER_SIZE = 8 * 1024 * 1024

// var TEST_FILE = IntputFile{"test_data/measurements_10.txt", 10}

//var TEST_FILE = IntputFile{"test_data/measurements_1k.txt", 1000}

// var TEST_FILE = IntputFile{"test_data/measurements_100k.txt", 100 * 1000}

// var TEST_FILE = IntputFile{"test_data/measurements_10M.txt", 10 * 1000 * 1000}

var TEST_FILE = IntputFile{"test_data/measurements_1B.txt", 1000 * 1000 * 1000}

type IntputFile struct {
	fileName string
	nbLines  int64
}

func main() {

	// f, err := os.Create("profiles/")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	panic(err)
	// }
	// defer pprof.StopCPUProfile()

	// slog.SetLogLoggerLevel(slog.LevelDebug.Level())

	startTime := time.Now()
	sortedCityTemperatures := processFile(TEST_FILE)
	endTime := time.Now()

	for _, cityTemperatures := range sortedCityTemperatures {
		fmt.Println(cityTemperatures)
	}

	fmt.Printf("Execution time is %v\n", endTime.Sub(startTime))
	checkResult(TEST_FILE, sortedCityTemperatures)
}

func processFile(inputFile IntputFile) []CityTemperatures {

	splitFileReader := NewSplitFileReader(inputFile.fileName, NUMBER_OF_READERS)
	reducer := NewReducer()
	mapsChan, _ := splitFileReader.processFileConcurrently()

	for i := 0; i < NUMBER_OF_READERS; i++ {
		partialResultMap := <-mapsChan
		reducer.reduce(partialResultMap)
	}

	sortedCityTemperatures := sorted(reducer.resultMap)
	return sortedCityTemperatures
}

func sorted(resultMap map[uint32]*CityTemperatures) []CityTemperatures {
	var resultSlice []CityTemperatures
	for _, ct := range resultMap {
		resultSlice = append(resultSlice, *ct)
	}
	slices.SortFunc(resultSlice, sortByName)
	return resultSlice
}

func checkResult(inputFile IntputFile, result []CityTemperatures) {
	var nbLinesProcessed int64
	for _, v := range result {
		nbLinesProcessed += int64(v.count)
	}
	if inputFile.nbLines != nbLinesProcessed {
		panic(fmt.Sprintf("Number of lines processed is incorrect ! Expected: %v, Got: %v", inputFile.nbLines, nbLinesProcessed))
	}
}
