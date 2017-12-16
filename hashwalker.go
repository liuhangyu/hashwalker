package main

import (
	"path/filepath"
	"os"
	"fmt"
	"time"
	"crypto/sha1"
	"bufio"
	"encoding/hex"
	"flag"
)

type HashWalker struct {
	dirpath string
	file *os.File
	filter []string
}

func CalcSha1(filepath string) (int64, string, string, error) {
	file, err := os.Open(filepath)
	if err!= nil {
		fmt.Println("failed to open")
		return 0,"","",err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0,"","",err
	}
	size := fileInfo.Size()
	fname := fileInfo.Name()


	reader := bufio.NewReader(file)
	context := make([]byte, size)
	sizebytes, err := reader.Read(context)
	if err != nil {
		return 0,"","",err
	}

	h := sha1.New()
	h.Write(context[:sizebytes])
	hash := h.Sum(nil)
	hxHash :=  hex.EncodeToString(hash)

	return size, fname, hxHash, nil
}

func (hw *HashWalker) OpenDir() ([]string, error){
	var dirList []string
	err := filepath.Walk(hw.dirpath,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if !f.IsDir() {
				dirList = append(dirList, path)
				return nil
			}
			return nil
		})
	return dirList, err
}

func checkFileIsExist(filename string) (bool) {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func (hw *HashWalker) CreateResultFile() error{
	var err error
	filename := hw.dirpath + "-" + time.Now().Format("2006-01-02-15-04-05")
	if checkFileIsExist(filename) {
		hw.file, err = os.OpenFile(filename, os.O_APPEND, 0600)
		if err != nil {
			fmt.Printf("%s filename is exist",filename)
		}
	} else {
		hw.file, err = os.Create(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func (hw *HashWalker) WriteResult(i int, size int64, filename string, fhash string) error {
	if hw.file == nil {
		fmt.Errorf("%s","file is not open")
	}
	writer := bufio.NewWriter(hw.file)
	result := fmt.Sprintf("{%d,%dbytes,%s,%s}\n", i, size, filename, fhash)
	fmt.Println(result)
	writer.WriteString(result)
	writer.Flush()
	return nil
}

func (hw *HashWalker) CloseResultFile() error{
	return hw.file.Close()
}

func (hw *HashWalker) CalcShaFile(fileList []string) error {
	for index, fpath := range fileList {
		//dir, file := filepath.Split(fpath)
		size, fname, fhash, err := CalcSha1(fpath)
		if err != nil {
			return err
		}
		fmt.Println(index, size, fname, fhash)
		err = hw.WriteResult(index, size, fname, fhash)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var dirpath string
	flag.StringVar(&dirpath, "dir", "./test", "dir path")
	flag.Parse()

	hw := HashWalker{dirpath: dirpath}

	err := hw.CreateResultFile()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	flist, err := hw.OpenDir()
	if err == nil {
		err := hw.CalcShaFile(flist)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	hw.CloseResultFile()
}
