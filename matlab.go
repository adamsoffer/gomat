package gomat

/*
#cgo CFLAGS: -I/Applications/MATLAB_R2014a.app/extern/include/
#cgo LDFLAGS: -L/Applications/MATLAB_R2014a.app/bin/maci64 -lmx -lmex -lmat
#include "/Applications/MATLAB_R2014a.app/extern/include/mat.h"
#include "/Applications/MATLAB_R2014a.app/extern/include/matrix.h"
#include "/Applications/MATLAB_R2014a.app/extern/include/mex.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type ChannelStatistics struct {
	minSignal float64
	maxSignal float64
}

// A Segment struct contains training data and correct classification state
type Segment struct {
	data [][]float64
}

func LoadMat(file string) interface{} {

	var pmat *C.MATFile
	var pa *C.mxArray
	var cs = C.CString(file)
	var r = C.CString("r")
	var name = C.CString("")

	// open file to get directory
	pmat = C.matOpen(cs, r)

	// get total directories of MAT-file
	x := C.int(0)
	C.matGetDir(pmat, &x)
	C.matClose(pmat)

	pmat = C.matOpen(cs, r)
	for i := 0; i < int(x); i++ {
		pa = C.matGetNextVariable(pmat, &name)
		totalFields := C.mxGetNumberOfFields(pa)

		// get each field in struct
		for x := 0; x < int(totalFields); x++ {
			var field *C.mxArray
			var realDataPtr *C.double

			fieldName := C.mxGetFieldNameByNumber(pa, C.int(x))

			fmt.Println(C.GoString(fieldName))

			if C.GoString(fieldName) == "data" {
				field = C.mxGetField(pa, 0, fieldName)
				rowTotal := C.mxGetM(field)
				colTotal := C.mxGetN(field)
				realDataPtr = C.mxGetPr(field)

				// Get max and min values for each row so we can normalize data
				channelStatistics := storeChannelStatistics(int(rowTotal), int(colTotal), realDataPtr)

				// Return normalized signal data inside two dimensional array
				data := getNormalizedMatrix(int(rowTotal), int(colTotal), realDataPtr, channelStatistics)

				fmt.Println(data[0][1])
				fmt.Println(data[0][2])
				fmt.Println(data[0][3])
				fmt.Println(data[0][4])
			}
		}
	}
	return 1
}

func getNormalizedMatrix(rowTotal int, colTotal int, realDataPtr *C.double, channelStatistics []*ChannelStatistics) [][]float64 {
	// Allocate the top-level slice.
	data := make([][]float64, int(rowTotal)) // One row per unit of y.
	// Loop over the rows, allocating the slice for each row.
	for i := range data {
		data[i] = make([]float64, int(colTotal))
	}
	// Normalize each element in the array and store in 2d slice
	for row := 0; row < rowTotal; row++ {
		ptr := uintptr(unsafe.Pointer(realDataPtr))
		// Go from row to row in memory (float is 8 bytes)
		ptr += 8 * uintptr(row)
		for col := 0; col < colTotal; col++ {

			signal := *(*float64)(unsafe.Pointer(ptr))
			minSignal := channelStatistics[row].minSignal
			maxSignal := channelStatistics[row].maxSignal
			value := normalize(signal, minSignal, maxSignal)
			data[row][col] = value
			// Matlab stores uses column major so we want to skip from
			// column to column using pointer arithmatic
			ptr += uintptr(rowTotal) * 8
		}
	}
	return data
}

func storeChannelStatistics(rowTotal int, colTotal int, realDataPtr *C.double) []*ChannelStatistics {
	var channelStatistics []*ChannelStatistics
	for row := 0; row < rowTotal; row++ {
		minSignal := 0.0
		maxSignal := 0.0
		ptr := uintptr(unsafe.Pointer(realDataPtr))
		// Go from row to row in memory (float is 8 bytes)
		ptr += 8 * uintptr(row)
		for col := 0; col < colTotal; col++ {
			signal := *(*float64)(unsafe.Pointer(ptr))
			// since it's column major we want to skip from
			// column to column using pointer arithmatic
			ptr += uintptr(rowTotal) * 8

			// store min value in row for normalization
			if signal < minSignal {
				minSignal = signal
			}
			// store max value in row for normalization
			if signal > maxSignal {
				maxSignal = signal
			}
		}
		// store min and max in channel slice
		channel := new(ChannelStatistics)
		channel.minSignal = minSignal
		channel.maxSignal = maxSignal
		channelStatistics = append(channelStatistics, channel)
	}
	return channelStatistics
}

// normalize value
func normalize(signal float64, minSignal float64, maxSignal float64) float64 {
	return (signal - minSignal) / (maxSignal - minSignal)
}
