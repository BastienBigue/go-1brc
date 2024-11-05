package main

import (
	"fmt"
	"math"
)

type MinMaxAverage struct {
	min, max, average float64
	count             int32
}

func (r *MinMaxAverage) String() string {
	return fmt.Sprintf("{min=%v ; max=%v ; average=%v}", r.min, r.max, r.average)
}

func (r *MinMaxAverage) updateWith(f float64) {
	r.max = math.Max(r.max, f)
	r.min = math.Min(r.min, f)
	r.average = (r.average*float64(r.count) + f) / float64(r.count+1)
	r.count++
}
