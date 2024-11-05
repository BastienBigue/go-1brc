package main

import "sync"

type SafeMap struct {
	citiesMap map[string]*MinMaxAverage
	mut       sync.Mutex
}

func (sm *SafeMap) exists(s string) bool {
	sm.mut.Lock()
	defer sm.mut.Unlock()
	_, ok := sm.citiesMap[s]
	return ok
}

func (sm *SafeMap) addOrUpdate(city string, temperature float64) {
	sm.mut.Lock()
	existingEntry, exists := sm.citiesMap[city]
	if !exists {
		sm.citiesMap[city] = &MinMaxAverage{min: temperature, max: temperature, count: 1, average: temperature}
	} else {
		existingEntry.updateWith(temperature)
	}
	sm.mut.Unlock()
}
