package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"
)

func main() {
	startTime := time.Now()
	readString(s)
	endTime := time.Now()
	fmt.Printf("Execution time is %v\n", endTime.Sub(startTime))
}

var s string = `Hamburg;12.0
Bulawayo;8.9
Palembang;38.8
St. John's;15.2
Cracow;12.6
Bridgetown;26.9
Istanbul;6.2
Roseau;34.4
Conakry;31.2
Istanbul;23.0`

func readString(s string) {
	reader := csv.NewReader(strings.NewReader(s))
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
		fmt.Println(record)
	}

}
