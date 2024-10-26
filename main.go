package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"time"
)

func main() {
	startTime := time.Now()
	processFile("temperatures.csv")
	endTime := time.Now()
	fmt.Printf("Execution time is %v\n", endTime.Sub(startTime))
}

type MinMaxAverage struct {
	min, max, average float64
	count             int32
}

func (r *MinMaxAverage) String() string {
	return fmt.Sprintf("{min=%v ; max=%v ; average=%v}", r.min, r.max, r.average)
}

var citiesMap map[string]*MinMaxAverage = make(map[string]*MinMaxAverage)

func processFile(s string) {
	f, _ := os.Open(s)
	reader := csv.NewReader(f)
	reader.Comma = ';'
	reader.Comment = '#'
	reader.FieldsPerRecord = 2
	//LazyQuotes = false
	//TrimLeadingSpace = true
	//ReuseRecord = false
	for {
		record, err := reader.Read()
		if err == io.EOF {
			fmt.Println("EOF")
			break
		}
		if err != nil {
			fmt.Println(err)
			break
		}
		processRecord(record)
		fmt.Println(record)
	}
	fmt.Println(citiesMap)
}

func processRecord(record []string) {
	city := record[0]
	temperature, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		fmt.Println(err)
		return
	}
	existingEntry, exists := citiesMap[city]
	if !exists {
		citiesMap[city] = &MinMaxAverage{min: temperature, max: temperature, count: 1, average: temperature}
	} else {
		existingEntry.updateWith(temperature)
	}

}

func (r *MinMaxAverage) updateWith(f float64) {
	r.max = math.Max(r.max, f)
	r.min = math.Min(r.min, f)
	r.average = (r.average*float64(r.count) + f) / float64(r.count+1)
	r.count++
}
