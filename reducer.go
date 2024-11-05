package main

import "log/slog"

type Reducer struct {
	resultMap map[string]*MinMaxAverage
}

func NewReducer() Reducer {
	return Reducer{resultMap: make(map[string]*MinMaxAverage)}
}

func (r *Reducer) reduce(partialResultMap map[string]*MinMaxAverage) {
	slog.Info("Reduce")
	for k1, v1 := range partialResultMap {
		v2, ok := r.resultMap[k1]
		if ok {
			r.resultMap[k1] = v2.mergeWith(v1)
		} else {
			r.resultMap[k1] = v1
		}
	}
}
