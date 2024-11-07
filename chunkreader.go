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

		dummyBytes := make([]byte, 105) //105 bytes long to be sure to have at least a full line
		_, err2 := f.ReadAt(dummyBytes, cr.from)
		if err2 != nil && err2 != io.EOF {
			slog.Error(err2.Error())
			return 0, err2
		}

		for i, currByte := range dummyBytes {
			if currByte == LineBreak {
				res = int64(i) + 1
				break
			}
		}
		return res, nil
	} else {
		return cr.from, nil
	}
}

func (cr *ChunkReader) startReader() error {
	slog.Info(fmt.Sprintf("Reader%v - Start startReader with chunkLength=%v", cr.readerNb, cr.chunkSize))
	bytesToSkipBecauseConsumedByPreviousReader, err := cr.bytesToSkipBecauseConsumedByPreviousReader()
	slog.Debug(fmt.Sprintf("Reader%v - Reader will skip %v bytes at the beginning of its chunk", cr.readerNb, bytesToSkipBecauseConsumedByPreviousReader))
	if err != nil {
		return err
	}
	nbBytesToRead := cr.chunkSize - bytesToSkipBecauseConsumedByPreviousReader
	for nbBytesToRead > 0 {

		slog.Debug(fmt.Sprintf("Reader%v - Still %v bytes to read", cr.readerNb, nbBytesToRead))

		f, err := os.Open(cr.fileName)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		defer f.Close()

		bytesSlice := make([]byte, READ_BUFFER_SIZE)
		readAtOffset := cr.from + cr.chunkSize - nbBytesToRead
		slog.Debug(fmt.Sprintf("Reader%v - Read from offset %v", cr.readerNb, readAtOffset))

		var byteBuffer ByteBuffer
		nbBytesRead, err2 := f.ReadAt(bytesSlice, readAtOffset)
		if err2 != nil && err2 != io.EOF {
			slog.Error(err.Error())
			return err
		} else if err2 == io.EOF {
			byteBuffer = ByteBuffer{byteBuffer: bytesSlice[:nbBytesRead], containsEOF: true}

			nbBytesToRead = int64(nbBytesRead)
			slog.Debug(fmt.Sprintf("Reader%v - Current buffer contains EOF. Its size is %v", cr.readerNb, nbBytesRead))
		} else {
			byteBuffer = ByteBuffer{byteBuffer: bytesSlice[:nbBytesRead], containsEOF: false}
		}
		slog.Debug(fmt.Sprintf("Reader%v - Buffer that will be processed : \n%v", cr.readerNb, string(bytesSlice)))

		nbBytesToRead = cr.processBuffer(byteBuffer, nbBytesToRead)

	}
	cr.chunkResultChan <- cr.chunkResultMap
	slog.Info(fmt.Sprintf("Reader%v - Reader%v is done!", cr.readerNb, cr.readerNb))
	return nil
}

type ByteBuffer struct {
	byteBuffer  []byte
	containsEOF bool
}

func (cr *ChunkReader) processBuffer(byteBuffer ByteBuffer, bytesToConsume int64) int64 {
	startLineOffset := 0
	temperatureOffset := 0
	var currentCity, temperature []byte
	for i, byteRead := range byteBuffer.byteBuffer {
		//slog.Debug(fmt.Sprintf("%v;%v\n", i, string(byteRead)))
		if byteRead == SemiColon {
			currentCity = byteBuffer.byteBuffer[startLineOffset:i]
			temperatureOffset = i + 1
		}
		if byteRead == LineBreak || (byteBuffer.containsEOF && i == len(byteBuffer.byteBuffer)-1) {
			//slog.Debug(fmt.Sprintf("i=%v;len(bytes)=%v\n", i, len(byteBuffer.byteBuffer)))
			//slog.Debug(fmt.Sprintf("Reader%v - Process line between offsets %v and %v\n", cr.readerNb, startLineOffset, i))
			temperature = byteBuffer.byteBuffer[temperatureOffset:i]
			cr.processRecord(currentCity, temperature)
			bytesToConsume -= int64(i - startLineOffset + 1) //+1 for line break we jump over
			if bytesToConsume < 0 {
				break
			}
			startLineOffset = i + 1
		}
	}
	return bytesToConsume
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
