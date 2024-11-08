package main

type Reducer struct {
	resultMap map[uint32]*CityTemperatures
}

func NewReducer() Reducer {
	return Reducer{resultMap: make(map[uint32]*CityTemperatures)}
}

func (r *Reducer) reduce(partialResultMap map[uint32]*CityTemperatures) {
	for k1, v1 := range partialResultMap {
		v2, ok := r.resultMap[k1]
		if ok {
			r.resultMap[k1] = v2.mergeWith(v1)
		} else {
			r.resultMap[k1] = v1
		}
	}
}
