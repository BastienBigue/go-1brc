package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
)

const SemiColon byte = byte(';')
const LineBreak byte = byte('\n')

type ChunkReader struct {
	fileName        string
	readerNb        uint8
	from, chunkSize int64
	chunkResultChan chan map[string]*MinMaxAverage
	chunkResultMap  map[string]*MinMaxAverage
}

func NewChunkReader(fileName string, readerNb uint8, from int64, chunkSize int64, chunkResultChan chan map[string]*MinMaxAverage) ChunkReader {
	cr := ChunkReader{
		fileName:        fileName,
		readerNb:        readerNb,
		chunkSize:       chunkSize,
		chunkResultChan: chunkResultChan,
		from:            from,
		chunkResultMap:  make(map[string]*MinMaxAverage),
	}
	slog.Debug("Create reader",
		"readerNb", cr.readerNb,
		"chunkSize", cr.chunkSize,
		"from", cr.from)
	return cr
}

func (cr *ChunkReader) atStartOfFile() bool {
	return cr.from == 0
}

func (cr *ChunkReader) bytesToSkipBecausePartOfPreviousChunk() int64 {
	//slog.Debug("bytesToSkipBecausePartOfPreviousChunk", "reader", cr.readerNb, "from", cr.from)

	if cr.atStartOfFile() {
		return cr.from
	} else {
		f, err := os.Open(cr.fileName)
		if err != nil {
			slog.Error(err.Error())
			panic(err)

		}
		defer f.Close()

		//106 bytes long to be sure to have at least one full line (100 bytes of city name + 1 byte separator + 4 bytes temperature + 1 byte LineBreak)
		dummyBytes := make([]byte, 106)

		// cr.from-1 to catch if cr.from is at start of line
		_, err2 := f.ReadAt(dummyBytes, cr.from-1)
		if err2 != nil && err2 != io.EOF {
			slog.Error(err2.Error())
			panic(err2)
		}

		// If true, means that cr.from is at start of line, and no need to skip any bytes
		if dummyBytes[0] == LineBreak {
			return 0
		}

		// Find next end of line, and return number of bytes that are part of an incomplete line at the start of the chunk
		var res int64
		for i, currByte := range dummyBytes {
			if currByte == LineBreak {
				res = int64(i) + 1
				break
			}
		}
		return res
	}
}

// Remove, if necessary, trailing bytes that are not part of this chunk, while keeping the last line complete
func (cr *ChunkReader) removeTrailingBytes(bytesBuffer []byte, nbBytesLeftInChunk int64, nbBytesInBuffer int) []byte {
	if nbBytesLeftInChunk < int64(nbBytesInBuffer) {
		//slog.Debug("startReader - remove trailing", "reader", cr.readerNb, "nbBytesLeftInChunk", nbBytesLeftInChunk, "nbBytesInBuffer", nbBytesInBuffer)
		for i, byte := range bytesBuffer[nbBytesLeftInChunk-1:] {
			if byte == LineBreak {
				bytesBuffer = bytesBuffer[:nbBytesLeftInChunk+int64(i)]
				break
			}
		}
		return bytesBuffer
	}
	return bytesBuffer
}

func (cr *ChunkReader) startReader() {
	slog.Info("startReader - start!", "reader", cr.readerNb)

	bytesToSkipBecauseConsumedByPreviousReader := cr.bytesToSkipBecausePartOfPreviousChunk()
	//slog.Debug("Skip bytes at start of chunk", "reader", cr.readerNb, "bytesSkipped", bytesToSkipBecauseConsumedByPreviousReader)

	// Reduces number of bytes to read by the number of bytes present at the beginning of the chunk and that were processed by the previous reader.
	nbBytesLeftInChunk := cr.chunkSize - bytesToSkipBecauseConsumedByPreviousReader

	for nbBytesLeftInChunk > 0 {

		// slog.Debug("startReader - loop", "reader", cr.readerNb, "bytes to read", nbBytesLeftInChunk)

		f, err := os.Open(cr.fileName)
		if err != nil {
			slog.Error(err.Error())
			panic(err)
		}
		defer f.Close()

		bytesBuffer := make([]byte, READ_BUFFER_SIZE)

		// Offset where we should read is chunk end - bytes read
		readAtOffset := cr.from + cr.chunkSize - nbBytesLeftInChunk
		// slog.Debug("startReader - read", "reader", cr.readerNb, "Read offset", readAtOffset)
		nbBytesInBuffer, err2 := f.ReadAt(bytesBuffer, readAtOffset)
		if err2 != nil && err2 != io.EOF {
			slog.Error(err.Error())
			panic(err2)
		}
		slog.Debug(fmt.Sprintf("Reader%v - Buffer that will be processed : \n%v", cr.readerNb, string(bytesSlice)))

		bytesBuffer = cr.removeTrailingBytes(bytesBuffer, nbBytesLeftInChunk, nbBytesInBuffer)
		nbBytesLeftInChunk -= cr.processBuffer(bytesBuffer)

	}
	cr.chunkResultChan <- cr.chunkResultMap
	slog.Info("startReader - done!", "reader", cr.readerNb)
}

func (cr *ChunkReader) processBuffer(byteBuffer []byte) int64 {
	//slog.Debug("processBuffer", "readerNb", cr.readerNb, "nbBytesLeftInChunk", nbBytesLeftInChunk, "len(byteBuffer)", len(byteBuffer), "byteBuffer", string(byteBuffer))
	startLineOffset := 0
	temperatureOffset := 0
	var currentCity, temperature []byte
	var byteProcessedInBuffer int64
	for currentOffset, byteRead := range byteBuffer {
		if byteRead == SemiColon {
			currentCity = byteBuffer[startLineOffset:currentOffset]
			temperatureOffset = currentOffset + 1
		}
		if byteRead == LineBreak {
			//slog.Debug("processBuffer - Line break found", "reader", cr.readerNb, "startLineOffset", startLineOffset, "endLineOffset", currentOffset, "line", byteBuffer[startLineOffset:currentOffset])

			// Removes "." from decimal
			temperature = append(byteBuffer[temperatureOffset:currentOffset-2], byteBuffer[currentOffset-1])

			cr.processRecord(currentCity, temperature)

			// Number of bytes processed is end-start+1
			byteProcessedInBuffer += int64(currentOffset - startLineOffset + 1)

			// Update startLineOffset for next line
			startLineOffset = currentOffset + 1
		}
	}
	return byteProcessedInBuffer
}

func (cr *ChunkReader) processRecord(city []byte, temperature []byte) {
	cityS := string(city)
	temperatureInt64, err := strconv.ParseInt(string(temperature), 10, 32)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
	temperatureInt32 := int32(temperatureInt64)

	//slog.Debug(fmt.Sprintf("Reader%v - Record : %v ; %v\n", cr.readerNb, cityS, temperatureF))
	existingEntry, exists := cr.chunkResultMap[cityS]
	if !exists {
		cr.chunkResultMap[cityS] = &MinMaxAverage{min: temperatureInt32, max: temperatureInt32, count: 1, sum: temperatureInt32}
	} else {
		existingEntry.updateWith(temperatureInt32)
	}
}
