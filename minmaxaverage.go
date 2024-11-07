package main

import (
	"fmt"
)

type MinMaxAverage struct {
	min, max, sum, count int32
}

func (mma *MinMaxAverage) String() string {
	return fmt.Sprintf("{min=%v ; max=%v ; average=%v ; sum=%v ; count=%v}", float64(mma.min)/10, float64(mma.max)/10, float64(mma.sum)/(float64(mma.count)*10), mma.sum, mma.count)
}

func (mma *MinMaxAverage) updateWith(f int32) {
	mma.max = max(mma.max, f)
	mma.min = min(mma.min, f)
	mma.sum = mma.sum + f
	mma.count++
}

func (mma *MinMaxAverage) mergeWith(other *MinMaxAverage) *MinMaxAverage {
	return &MinMaxAverage{
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
