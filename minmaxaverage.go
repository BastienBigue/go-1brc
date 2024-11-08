package main

import (
	"fmt"
)

type CityTemperatures struct {
	city                 []byte
	min, max, sum, count int32
}

func (mma *CityTemperatures) String() string {
	return fmt.Sprintf("%v : {min=%v ; max=%v ; average=%v ; sum=%v ; count=%v}", string(mma.city), float64(mma.min)/10, float64(mma.max)/10, float64(mma.sum)/(float64(mma.count)*10), mma.sum, mma.count)
}

func (mma *CityTemperatures) updateWith(f int32) {
	mma.max = max(mma.max, f)
	mma.min = min(mma.min, f)
	mma.sum = mma.sum + f
	mma.count++
}

func (mma *CityTemperatures) mergeWith(other *CityTemperatures) *CityTemperatures {
	return &CityTemperatures{
		max:   max(mma.max, other.max),
		min:   min(mma.min, other.min),
		sum:   mma.sum + other.sum,
		count: mma.count + other.count}
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
