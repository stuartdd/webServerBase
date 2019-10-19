package largefile

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/stuartdd/webServerBase/panicapi"
)

/*
LargeFileData contains the data associated with a large file.
*/
type LargeFileData struct {
	Name      string
	Offsets   []int64
	LineCount int
	Size      int64
	Time      time.Time
	bufSize   int
	extendBy  int
}

var pageMap = make(map[string]*LargeFileData)

/*
 */
func NewLargeFileReader(name string) *LargeFileData {
	return NewLargeFileReaderDetailed(name, 100, 100, 50)
}

/*
NewLargeFileReaderDetailed Initialise the large File Reader data set
*/
func NewLargeFileReaderDetailed(name string, fileReaderBufferSize int, initialLineCount int, extendBy int) *LargeFileData {
	if fileReaderBufferSize == 0 {
		panicapi.ThrowError(500, panicapi.SCParamValidation, "Internal Server Error", "NewLargeFileReader: Internal error: openInitial-->fileReaderBufferSize Parameter cannot be 0")
	}
	info, err := os.Stat(name)
	if err != nil {
		panicapi.ThrowError(404, panicapi.SCFileNotFound, "Not Found", fmt.Sprintf("NewLargeFileReader: File %s could not be found. %s", name, err.Error()))
	}
	f, err := os.Open(name)
	if err != nil {
		panicapi.ThrowError(417, panicapi.SCOpenFileError, "Expectation Failed", fmt.Sprintf("NewLargeFileReader: File %s could not be opened. %s", name, err.Error()))
	}
	defer f.Close()

	var offset int64                          // Initial offset in to the file!
	bytesRead := 0                            // The number of bytest read
	notEOF := true                            // Are we at the end of the file
	buf := make([]byte, fileReaderBufferSize) // Buffer for the file contents
	/*
		Init the data structure
	*/
	data := &LargeFileData{
		Name:      name,
		Offsets:   make([]int64, initialLineCount), // Make room for 100 lines
		LineCount: 0,
		Size:      info.Size(),          // The size of the file so we know if it is extended
		Time:      time.Now(),           // The time we last read the file so we can clean up later
		bufSize:   fileReaderBufferSize, // Keep this for ReadMoreLines to use
		extendBy:  extendBy,             // Extend the offsets array by this amount
	}
	/*
		While not at the end of the file
	*/
	for notEOF {
		/*
			Read a buffer sized chunk
		*/
		bytesRead, err = io.ReadAtLeast(f, buf, data.bufSize)
		/*
			Check for errors and End Of File
		*/
		notEOF = checkOpenInitialError(name, "NewLargeFileReader", err)
		/*
			Parse the buffer for line feeds and record their position
		*/
		offset = data.parseOpenInitial(bytesRead, buf, offset)
	}
	/*
		Add an empty line to the end of the file so me now how big the file is
	*/
	offset = data.parseOpenInitial(2, []byte{10, 32}, offset)
	/*
		Add the data to the map so we can get it back
	*/
	pageMap[name] = data
	return data
}

/*
ReadLargeFile Read 'count' lines from the file from line 'from'
*/
func (p *LargeFileData) ReadLargeFile(from int, count int) string {

	info, err := os.Stat(p.Name)
	if err != nil {
		panicapi.ThrowError(404, panicapi.SCFileNotFound, "Not Found", fmt.Sprintf("File %s could not be found. %s", p.Name, err.Error()))
	}
	if count == 0 {
		return ""
	}
	to := from + count

	if to > p.LineCount {
		if info.Size() != p.Size {
			/*
				File has changed
			*/
			p.readMoreLines()
		}
	}
	if to < from {
		return ""
	}

	/*
		If from is 0 then read from the start!
	*/
	var start int64
	if from > 0 {
		start = p.Offsets[from-1] + 1 // To skip from the new line char to the start of the line!
	}

	end := p.Offsets[p.LineCount-1]
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
		panicapi.ThrowError(417, panicapi.SCOpenFileError, "Expectation Failed", fmt.Sprintf("File %s could not be opened. %s", p.Name, err.Error()))
	}
	defer f.Close()
	/*
		Read from a point in the file
	*/
	if start > 0 {
		_, err = f.Seek(start, 0)
		if err != nil {
			panicapi.ThrowError(417, panicapi.SCOpenFileError, "Expectation Failed", fmt.Sprintf("File %s could not seek. %s", p.Name, err.Error()))
		}
	}

	/*
		Read the rquired number of bytes
	*/
	bytes, err := io.ReadAtLeast(f, buf, int(bytesToRead))
	checkOpenInitialError(p.Name, "ReadLargeFile", err)

	if bytes < 1 {
		return ""
	}
	return string(buf[0:bytes])
}

/*
GetLargeFileReader returns LargeFileReader data for a specific file.
*/
func GetLargeFileReader(name string) *LargeFileData {
	lfr := pageMap[name]
	if lfr == nil {
		return nil
	}
	return lfr
}

func (p *LargeFileData) readMoreLines() {
	info, err := os.Stat(p.Name)
	if err != nil {
		panicapi.ThrowError(404, panicapi.SCFileNotFound, "Not Found", fmt.Sprintf("ReadMoreLines: File %s could not be found. %s", p.Name, err.Error()))
	}
	f, err := os.Open(p.Name)
	if err != nil {
		panicapi.ThrowError(417, panicapi.SCOpenFileError, "Expectation Failed", fmt.Sprintf("ReadMoreLines: File %s could not be opened. %s", p.Name, err.Error()))
	}
	defer f.Close()

	offset := p.Offsets[p.LineCount-1] // Offset in to the file!
	bytesRead := 0                     // The number of bytest read
	notEOF := true                     // Are we at the end of the file
	buf := make([]byte, p.bufSize)     // Buffer for the file

	/*
		Pluss 1 so it is at the start of the next line!
	*/
	offset, err = f.Seek(offset+1, 0)
	if err != nil {
		panicapi.ThrowError(417, panicapi.SCOpenFileError, "Expectation Failed", fmt.Sprintf("ReadMoreLines: File %s could not seek. %s", p.Name, err.Error()))
	}

	p.Size = info.Size()
	p.Time = time.Now()

	for notEOF {
		bytesRead, err = io.ReadAtLeast(f, buf, p.bufSize)
		notEOF = checkOpenInitialError(p.Name, "NewLargeFileReader", err)
		offset = p.parseOpenInitial(bytesRead, buf, offset)
	}
	/*
		Add an empty line to the end of the file so me now how big the file is
	*/
	offset = p.parseOpenInitial(2, []byte{10, 32}, offset)
}

/*
for each new line add the offset in the file to that line in the offsets
*/
func (p *LargeFileData) parseOpenInitial(bytesRead int, b []byte, offset int64) int64 {
	for i := 0; i < bytesRead; i++ {
		if b[i] == 10 {
			if p.LineCount >= len(p.Offsets) {
				newLen := p.LineCount + p.extendBy
				sb := make([]int64, newLen)
				copy(sb, p.Offsets)
				p.Offsets = sb
			}
			p.Offsets[p.LineCount] = offset
			p.LineCount++
		}
		offset++
	}
	return offset
}

func checkOpenInitialError(name string, context string, err error) bool {
	if err != nil {
		switch err {
		case io.EOF, io.ErrUnexpectedEOF:
			return false
		default:
			panicapi.ThrowError(400, panicapi.SCOpenFileError, "Read File", fmt.Sprintf("%s: File %s could not read. %s", context, name, err.Error()))
		}
	}
	return true
}
