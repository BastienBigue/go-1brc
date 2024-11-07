package main

import (
	"fmt"
	"log/slog"
	"os"
)

type SplitFileReaders struct {
	fileName  string
	nbReaders int
	fileSize  int64
}

func NewSplitFileReader(fileName string, nbReaders int) SplitFileReaders {
	return SplitFileReaders{
		fileName:  fileName,
		nbReaders: nbReaders,
		fileSize:  fileSize(fileName)}
}

func (sfr *SplitFileReaders) processFileConcurrently() (chan map[string]*MinMaxAverage, error) {
	mapsChan := make(chan map[string]*MinMaxAverage)
	slog.Info(fmt.Sprintf("Start %v readers for file %v", sfr.nbReaders, sfr.fileName))
	sfr.startReaders(mapsChan)
	return mapsChan, nil
}

func fileSize(fileName string) int64 {
	stat, err := os.Stat(fileName)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
	return stat.Size()
}

func (sfr *SplitFileReaders) defaultChunkSize() int64 {
	return sfr.fileSize/int64(sfr.nbReaders) + 1
}

func (sfr *SplitFileReaders) chunkSizeForReader(readerNb int) int64 {
	if readerNb == sfr.nbReaders-1 {
		return sfr.fileSize - (int64(sfr.nbReaders)-1)*sfr.defaultChunkSize()
	} else {
		return sfr.defaultChunkSize()
	}
}

func (sfr *SplitFileReaders) startReaders(chunkResultChan chan map[string]*MinMaxAverage) {
	for i := 0; i < sfr.nbReaders; i++ {
		r := NewChunkReader(sfr.fileName, uint8(i), int64(i)*sfr.defaultChunkSize(), sfr.chunkSizeForReader(i), chunkResultChan)
		go r.startReader()
	}
}
