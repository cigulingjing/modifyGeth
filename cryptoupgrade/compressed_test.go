package cryptoupgrade

import (
	"fmt"
	"os"
	"testing"

)

// * The test will create a copy in current folder.
func TestCompress(t *testing.T) {
	compressFile := "./compressed.go"
	compressedString, err := compressFileToString(compressFile)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	// rootpath is the path of the file to be executed
	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("get current workfolder err:%v\n", err)
		return
	} else {
		fmt.Printf("currentPath: %v\n", rootPath)
	}
	ouputPath := rootPath + "/decompressed" + "_" + "main.go"
	err = decompressStringToFile(compressedString, ouputPath)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
}

func TestAddress(t *testing.T) {
	// 0x0000000000000000000000000000000000000043
	fmt.Printf("CodeStorageAddress: %s\n", CodeStorageAddress)
	// 0x4f6c549854117e07a5568cf1a0065a0214f99bf28a89e0a839c88e7329de7fb0
	fmt.Printf("PullCodeEventHash: %v\n", PullCodeEventHash)
}
