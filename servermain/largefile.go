package servermain

import (
	"fmt"
	"io"
	"os"
	"time"
)

type largeFileData struct {
	Name      string
	Offsets   []int64
	LineCount int
	ModTime   time.Time
	Time      time.Time
}

var pageMap = make(map[string]*largeFileData)

func (p *largeFileData) ReadLargeFile(from int, count int) string {

	info, err := os.Stat(p.Name)
	if err != nil {
		ThrowPanic("E", 404, SCFileNotFound, "Not Found", fmt.Sprintf("File %s could not be found. %s", p.Name, err.Error()))
	}
	if count == 0 {
		return ""
	}
	to := from + count

	if to >= p.LineCount {
		if info.ModTime().After(p.ModTime) {
			/*
				File has changed
			*/
		}
	}
	if to < from {
		return ""
	}

	/*
		If from is 0 then read from the start!
	*/
	var start int64 = 0
	if from > 0 {
		start = p.Offsets[from-1] + 1 // To skip from the new line char to the start of the line!
	}

	var end int64 = p.Offsets[p.LineCount-1]
	if to <= p.LineCount {
		end = p.Offsets[to-1]
	}

	bytesToRead := (end - start) + 1
	if bytesToRead < 1 {
		return ""
	}
	buf := make([]byte, bytesToRead)

	f, err := os.Open(p.Name)
	if err != nil {
		ThrowPanic("E", 417, SCOpenFileError, "Expectation Failed", fmt.Sprintf("File %s could not be opened. %s", p.Name, err.Error()))
	}
	defer f.Close()
	/*
		Read from a point in the file
	*/
	if start > 0 {
		_, err = f.Seek(start, 0)
		if err != nil {
			ThrowPanic("E", 417, SCOpenFileError, "Expectation Failed", fmt.Sprintf("File %s could not seek. %s", p.Name, err.Error()))
		}
	}

	/*
		Read the rquired number of bytes
	*/
	bytes, err := io.ReadAtLeast(f, buf, int(bytesToRead))
	checkOpenInitialError(p.Name, err)

	if bytes < 1 {
		return ""
	}
	return string(buf[0:bytes])
}

func GetLargeFileReader(name string) *largeFileData {
	lfr := pageMap[name]
	if lfr == nil {
		return nil
	}
	return lfr
}

func NewLargeFileReader(name string, fileReaderBufferSize int) *largeFileData {
	if fileReaderBufferSize == 0 {
		ThrowPanic("E", 500, SCParamValidation, "Internal Server Error", "Internal error: openInitial-->fileReaderBufferSize Parameter cannot be 0")
	}
	info, err := os.Stat(name)
	if err != nil {
		ThrowPanic("E", 404, SCFileNotFound, "Not Found", fmt.Sprintf("File %s could not be found. %s", name, err.Error()))
	}
	f, err := os.Open(name)
	if err != nil {
		ThrowPanic("E", 417, SCOpenFileError, "Expectation Failed", fmt.Sprintf("File %s could not be opened. %s", name, err.Error()))
	}
	defer f.Close()

	var offset int64 = 0                      // Offset in to the file!
	bytesRead := 0                            // The number of bytest read
	notEOF := true                            // Are we at the end of the file
	buf := make([]byte, fileReaderBufferSize) // Buffer for the file

	data := &largeFileData{
		Name:      name,
		Offsets:   make([]int64, 50), // Make room for 100 lines
		LineCount: 0,
		ModTime:   info.ModTime(), // The time the file was updated
		Time:      time.Now(),     // The time we read the file
	}
	/*
		While not at the end of the file
	*/
	for notEOF {
		bytesRead, err = io.ReadAtLeast(f, buf, fileReaderBufferSize)
		notEOF = checkOpenInitialError(name, err)
		offset = data.parseOpenInitial(bytesRead, buf, offset)
	}
	/*
		Add an empty line to the end of the file so me now how big the file is
	*/
	offset = data.parseOpenInitial(2, []byte{10, 32}, offset)
	pageMap[name] = data
	return data
}

/*
for each new line add the offset in the file to that line in the offsets
*/
func (p *largeFileData) parseOpenInitial(bytesRead int, b []byte, offset int64) int64 {
	for i := 0; i < bytesRead; i++ {
		if b[i] == 10 {
			if p.LineCount >= len(p.Offsets) {
				newLen := p.LineCount + 50
				sb := make([]int64, newLen)
				for i := 0; i < p.LineCount; i++ {
					sb[i] = p.Offsets[i]
				}
				p.Offsets = sb
			}
			p.Offsets[p.LineCount] = offset
			p.LineCount++
		}
		offset++
	}
	return offset
}

func checkOpenInitialError(name string, err error) bool {
	if err != nil {
		switch err {
		case io.EOF, io.ErrUnexpectedEOF:
			return false
		default:
			ThrowPanic("E", 400, SCOpenFileError, "Read File", fmt.Sprintf("File %s could not read. %s", name, err.Error()))
		}
	}
	return true
}