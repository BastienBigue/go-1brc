package main

import (
	"fmt"
	"log/slog"
	"os"
)

type SplitFileReaders struct {
	fileName  string
	nbReaders int
}

func (sfr *SplitFileReaders) processFileConcurrently() (chan map[string]*MinMaxAverage, error) {
	splitSize, err := sfr.splitSize()
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	mapsChan := make(chan map[string]*MinMaxAverage)
	slog.Info(fmt.Sprintf("Start %v readers for file %v", sfr.nbReaders, sfr.fileName))
	sfr.startReaders(splitSize, mapsChan)
	return mapsChan, nil
}

func (sfr *SplitFileReaders) splitSize() (int64, error) {
	f, err := os.Open(sfr.fileName)
	if err != nil {
		slog.Error(err.Error())
		return 0, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		slog.Error(err.Error())
		return 0, err
	}
	fileSize := stat.Size()
	splitSize := fileSize/int64(sfr.nbReaders) + 1
	return splitSize, nil
}

func (sfr *SplitFileReaders) startReaders(chunkSize int64, chunkResultChan chan map[string]*MinMaxAverage) {
	for i := 0; i < sfr.nbReaders; i++ {
		slog.Debug("Start reader", "from", int64(i)*chunkSize, "chunkLength", chunkSize)
		r := NewChunkReader(sfr.fileName, int64(i), chunkSize, chunkResultChan)
		go r.startReader()
	}
}
