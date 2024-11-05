package main

import (
	"fmt"
	"math"
)

type MinMaxAverage struct {
	min, max, average float64
	count             int32
}

func (mma *MinMaxAverage) String() string {
	return fmt.Sprintf("{min=%v ; max=%v ; average=%v ; count=%v}", mma.min, mma.max, mma.average, mma.count)
}

func (mma *MinMaxAverage) updateWith(f float64) {
	mma.max = math.Max(mma.max, f)
	mma.min = math.Min(mma.min, f)
	mma.average = (mma.average*float64(mma.count) + f) / float64(mma.count+1)
	mma.count++
}

func (mma *MinMaxAverage) mergeWith(other *MinMaxAverage) *MinMaxAverage {
	return &MinMaxAverage{
		max:     math.Max(mma.max, other.max),
		min:     math.Min(mma.min, other.min),
		average: (mma.average*float64(mma.count) + other.average*float64(other.count)) / float64(mma.count+other.count),
		count:   mma.count + other.count}
}
