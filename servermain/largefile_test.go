package servermain

import (
	"testing"

	"webServerExample.go/test"
)

const testFilePrefix = "../site/TestLargeFileRead-"

/* Offsets
0	1	2	3	4	5	6
6	13	20*					Actual file offsets *=EOF
6	13	20	0	0	0	0 	Initial state
6	13	20	27	34*			Actual file offests *=EOF
6	13	20	27	34	0	0 	After append

Chars per line is 7
*/
func TestExtendingFile(t *testing.T) {
	name := "tempFile.txt"
	test.DeleteFile(t, name, false)
	test.AppendToFile(t, name, "Line 1\nLine 2\nLine 3")
	defer test.DeleteFile(t, name, true)
	list := NewLargeFileReader(name, 50)

	lc := list.LineCount
	test.AssertIntEqual(t, "lc=5a", lc, 3)

	test.AssertStringEqualsUnix(t, "1=", list.ReadLargeFile(0, 1), "Line 1\n")
	test.AssertStringEqualsUnix(t, "2=", list.ReadLargeFile(1, 1), "Line 2\n")
	test.AssertStringEqualsUnix(t, "3=", list.ReadLargeFile(2, 1), "Line 3")
	test.AssertStringEqualsUnix(t, "4=", list.ReadLargeFile(3, 1), "")
	test.AssertIntEqual(t, "lc=5b", list.LineCount, lc)

	test.AppendToFile(t, name, "\nLine 4\nLine 5")
	test.AssertFileContains(t, "file", name, "Line 4", "Line 5")

	test.AssertStringEqualsUnix(t, "1=", list.ReadLargeFile(0, 1), "Line 1\n")
	test.AssertIntEqual(t, "", list.LineCount, lc)
	test.AssertStringEqualsUnix(t, "2+", list.ReadLargeFile(1, 1), "Line 2\n")
	test.AssertIntEqual(t, "", list.LineCount, lc)
	test.AssertStringEqualsUnix(t, "3+", list.ReadLargeFile(2, 1), "Line 3\n")
	test.AssertIntEqual(t, "", list.LineCount, lc)
	test.AssertStringEqualsUnix(t, "4+", list.ReadLargeFile(3, 1), "Line 4\n")
	test.AssertIntEqual(t, "", list.LineCount, lc+2)
	test.AssertStringEqualsUnix(t, "5+", list.ReadLargeFile(4, 1), "Line 5")
	test.AssertIntEqual(t, "", list.LineCount, lc+2)
	test.AssertStringEqualsUnix(t, "6+", list.ReadLargeFile(5, 1), "")
	test.AssertIntEqual(t, "", list.LineCount, lc+2)
}

/* ../site/TestLargeFileRead-002.txt
0: \n
1: b\n
2: c\n
3: \n
4: d
*/
func TestLargeFileRead003_RandomRead(t *testing.T) {
	name := testFilePrefix + "003.txt"
	list := NewLargeFileReader(name, 5)
	test.AssertStringContains(t, "15", list.ReadLargeFile(15, 1), "port", "8080")
	test.AssertStringContains(t, "27", list.ReadLargeFile(27, 1), "linux")
	test.AssertStringContains(t, "1", list.ReadLargeFile(1, 1), "loggerLevels")
	test.AssertStringContains(t, "3,3", list.ReadLargeFile(3, 3), "windows", "linux", "darwin")
}

func TestLargeFileRead002_1(t *testing.T) {
	name := testFilePrefix + "002.txt"
	list := NewLargeFileReader(name, 5)
	test.AssertStringEqualsUnix(t, "7", list.ReadLargeFile(0, 7), "\nb\nc\n\nd")
	test.AssertStringEqualsUnix(t, "6", list.ReadLargeFile(0, 6), "\nb\nc\n\nd")
	test.AssertStringEqualsUnix(t, "5", list.ReadLargeFile(0, 5), "\nb\nc\n\nd")
	test.AssertStringEqualsUnix(t, "4", list.ReadLargeFile(0, 4), "\nb\nc\n\n")
	test.AssertStringEqualsUnix(t, "3", list.ReadLargeFile(0, 3), "\nb\nc\n")
	test.AssertStringEqualsUnix(t, "2", list.ReadLargeFile(0, 2), "\nb\n")
	test.AssertStringEqualsUnix(t, "1", list.ReadLargeFile(0, 1), "\n")
	test.AssertStringEqualsUnix(t, "0", list.ReadLargeFile(0, 0), "")
}

/* ../site/TestLargeFileRead-001.txt
0: a\n
1: b\n
2: c\n
3: \n
4: d\n
*/
func TestLargeFileRead5(t *testing.T) {
	name := testFilePrefix + "001.txt"
	list := NewLargeFileReader(name, 20)
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(6, 3), "")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(5, 3), "")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(5, 2), "")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(5, 1), "")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(5, 0), "")
}

func TestLargeFileRead4(t *testing.T) {
	name := testFilePrefix + "001.txt"
	list := NewLargeFileReader(name, 5)
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(4, 4), "d\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(4, 3), "d\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(4, 2), "d\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(4, 1), "d\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(4, 0), "")
}

func TestLargeFileRead3(t *testing.T) {
	name := testFilePrefix + "001.txt"
	list := NewLargeFileReader(name, 5)
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(3, 99), "\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(3, 3), "\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(3, 2), "\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(3, 1), "\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(3, 0), "")
}

func TestLargeFileRead1(t *testing.T) {
	name := testFilePrefix + "001.txt"
	list := NewLargeFileReader(name, 5)
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(1, 99), "b\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(1, 5), "b\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(1, 4), "b\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(1, 3), "b\nc\n\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(1, 2), "b\nc\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(1, 1), "b\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(1, 0), "")
}

func TestLargeFileRead0(t *testing.T) {
	name := testFilePrefix + "001.txt"
	list := NewLargeFileReader(name, 5)
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(0, 99), "a\nb\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(0, 7), "a\nb\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(0, 6), "a\nb\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(0, 5), "a\nb\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(0, 4), "a\nb\nc\n\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(0, 3), "a\nb\nc\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(0, 2), "a\nb\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(0, 1), "a\n")
	test.AssertStringEqualsUnix(t, "", list.ReadLargeFile(0, 0), "")
}

// func TestLargeParse(t *testing.T) {
// 	name := testFilePrefix+"001.txt"
// 	list := NewLargeFileReader(name, 5)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 1)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 2)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 3)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 4)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 6)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 7)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 8)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 9)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 10)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 11)
// 	checkLFR(t, list)
// 	list = NewLargeFileReader(name, 101)
// 	checkLFR(t, list)
// }

// func checkLFR(t *testing.T, l *largeFileData) {
// 	test.AssertIntEqual(t, "", l.LineCount, 7)
// 	test.AssertInt64Equal(t, "", l.Offsets[0], 0)
// 	test.AssertInt64Equal(t, "", l.Offsets[1], 1)
// 	test.AssertInt64Equal(t, "", l.Offsets[2], 3)
// 	test.AssertInt64Equal(t, "", l.Offsets[3], 5)
// 	test.AssertInt64Equal(t, "", l.Offsets[4], 6)
// 	test.AssertInt64Equal(t, "", l.Offsets[5], 8)
// 	test.AssertInt64Equal(t, "", l.Offsets[6], 9)
// }
// 	test.AssertInt64Equal(t, "", l.Offsets[2], 3)
// 	test.AssertInt64Equal(t, "", l.Offsets[3], 5)
// 	test.AssertInt64Equal(t, "", l.Offsets[4], 6)
// 	test.AssertInt64Equal(t, "", l.Offsets[5], 8)
// 	test.AssertInt64Equal(t, "", l.Offsets[6], 9)
// }
// 	test.AssertInt64Equal(t, "", l.Offsets[4], 6)
// 	test.AssertInt64Equal(t, "", l.Offsets[5], 8)
// 	test.AssertInt64Equal(t, "", l.Offsets[6], 9)
// }
