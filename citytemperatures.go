package main

import (
	"fmt"
	"strings"
)

type CityTemperatures struct {
	city                 []byte
	min, max, sum, count int32
}

func (ct CityTemperatures) String() string {
	return fmt.Sprintf("%v : {min=%v ; max=%v ; average=%v ; sum=%v ; count=%v}", string(ct.city), float64(ct.min)/10, float64(ct.max)/10, float64(ct.sum)/(float64(ct.count)*10), ct.sum, ct.count)
}

func NewCityTemperatures(city []byte, temperature int32) *CityTemperatures {
	ct := CityTemperatures{city: city, min: temperature, max: temperature, sum: temperature, count: 1}
	return &ct
}

func (ct *CityTemperatures) updateWith(f int32) {
	ct.max = max(ct.max, f)
	ct.min = min(ct.min, f)
	ct.sum = ct.sum + f
	ct.count++
}

func (ct *CityTemperatures) mergeWith(other *CityTemperatures) *CityTemperatures {
	return &CityTemperatures{
		city:  ct.city,
		max:   max(ct.max, other.max),
		min:   min(ct.min, other.min),
		sum:   ct.sum + other.sum,
		count: ct.count + other.count}
}

func sortByName(ct, other CityTemperatures) int {
	return strings.Compare(string(ct.city), string(other.city))
}

func min(a int32, b int32) int32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func max(a int32, b int32) int32 {
	if a > b {
		return a
	} else {
		return b
	}
}
