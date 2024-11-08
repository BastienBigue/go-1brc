package main

import (
	"hash/fnv"
	"io"
	"log/slog"
	"os"
)

const SemiColon byte = byte(';')
const LineBreak byte = byte('\n')
const Minus byte = byte('-')

type ChunkReader struct {
	fileName        string
	readerNb        uint8
	from, chunkSize int64
	chunkResultChan chan map[uint32]*CityTemperatures
	chunkResultMap  map[uint32]*CityTemperatures
}

func NewChunkReader(fileName string, readerNb uint8, from int64, chunkSize int64, chunkResultChan chan map[uint32]*CityTemperatures) ChunkReader {
	cr := ChunkReader{
		fileName:        fileName,
		readerNb:        readerNb,
		chunkSize:       chunkSize,
		chunkResultChan: chunkResultChan,
		from:            from,
		chunkResultMap:  make(map[uint32]*CityTemperatures),
	}
	// slog.Debug("Create reader", "readerNb", cr.readerNb,	"chunkSize", cr.chunkSize, "from", cr.from)
	return cr
}

func (cr *ChunkReader) atStartOfFile() bool {
	return cr.from == 0
}

func (cr *ChunkReader) bytesToSkipBecausePartOfPreviousChunk() int64 {
	// slog.Debug("bytesToSkipBecausePartOfPreviousChunk", "reader", cr.readerNb, "from", cr.from)

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
				res = int64(i)
				break
			}
		}
		return res
	}
}

// Remove, if necessary, trailing bytes that are not part of this chunk, while keeping the last line complete
func (cr *ChunkReader) removeTrailingBytes(bytesBuffer []byte, nbBytesLeftInChunk int64, nbBytesInBuffer int) []byte {
	// slog.Debug("startReader - remove trailing", "reader", cr.readerNb, "nbBytesLeftInChunk", nbBytesLeftInChunk, "nbBytesInBuffer", nbBytesInBuffer)
	if nbBytesLeftInChunk < int64(nbBytesInBuffer) {
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
	slog.Info("startReader - start", "reader", cr.readerNb)

	bytesToSkipBecauseConsumedByPreviousReader := cr.bytesToSkipBecausePartOfPreviousChunk()
	// slog.Debug("Skip bytes at start of chunk", "reader", cr.readerNb, "bytesSkipped", bytesToSkipBecauseConsumedByPreviousReader)

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

		bytesBuffer = cr.removeTrailingBytes(bytesBuffer, nbBytesLeftInChunk, nbBytesInBuffer)
		nbBytesLeftInChunk -= cr.processBuffer(bytesBuffer)

	}
	cr.chunkResultChan <- cr.chunkResultMap
	slog.Info("startReader - done", "reader", cr.readerNb)
}

func (cr *ChunkReader) processBuffer(byteBuffer []byte) int64 {
	// slog.Debug("processBuffer", "reader", cr.readerNb, "len(byteBuffer)", len(byteBuffer), "byteBuffer", string(byteBuffer))
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
			// slog.Debug("processBuffer - Line break found", "reader", cr.readerNb, "startLineOffset", startLineOffset, "endLineOffset", currentOffset, "line", byteBuffer[startLineOffset:currentOffset])

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

	hash := fnv.New32()
	hash.Write(city)
	cityHash := hash.Sum32()

	temperatureInt32 := parseTemperatureAsInt(temperature)

	// slog.Debug("processRecord", "reader", cr.readerNb, "city", string(city), "temperature", temperatureInt32)
	existingEntry, exists := cr.chunkResultMap[cityHash]
	if !exists {
		cr.chunkResultMap[cityHash] = NewCityTemperatures(city, temperatureInt32)
	} else {
		existingEntry.updateWith(temperatureInt32)
	}
}

func parseTemperatureAsInt(temperatureBytes []byte) int32 {
	if len(temperatureBytes) == 4 { // temp < -10
		return -parseAbsTemperatureAsInt(temperatureBytes[1:])
	} else if len(temperatureBytes) == 3 && temperatureBytes[0] == Minus { // -10 < temp < 0
		return -parseAbsTemperatureAsInt(temperatureBytes[1:])
	} else if len(temperatureBytes) == 3 && temperatureBytes[0] != Minus { // temp > 10
		return parseAbsTemperatureAsInt(temperatureBytes)
	} else { // 0 <= temp < 10
		return parseAbsTemperatureAsInt(temperatureBytes)
	}
}

func parseAbsTemperatureAsInt(temperatureBytes []byte) int32 {
	if len(temperatureBytes) == 3 {
		return (int32(temperatureBytes[0])-48)*100 + (int32(temperatureBytes[1])-48)*10 + (int32(temperatureBytes[2]) - 48)
	} else if len(temperatureBytes) == 2 {
		return (int32(temperatureBytes[0])-48)*10 + (int32(temperatureBytes[1]) - 48)
	} else {
		panic("Temperature with less than 2 or more than 3 digits")
	}
}
