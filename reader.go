package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func ReadFile(graphId string, fileToRead string, maxLines int, startPos int64, re string, configName string, ch chan<- DataResp) {

	var resp DataResp
	var dataFloats []float64
	r := regexp.MustCompile(re)

	file, err := os.Open(fileToRead)
	if err != nil {
		// resp.ConfigName, resp.Data, resp.FileName, resp.Position, resp.Error = configName, dataFloats, fileToRead, startPos, err
		_, resp.Error = "", err
		ch <- resp
		return
	}
	defer file.Close()

	fileReader := bufio.NewReader(file)

	// set position to start reading
	file.Seek(startPos, io.SeekStart)
	for i := 0; i < maxLines; i++ {
		line, err := fileReader.ReadString('\n')
		if err != nil {
			// check if file does not exist anymore
			if _, err := os.Stat(fileToRead); errors.Is(err, os.ErrNotExist) {
				_, resp.Error = "", err
				ch <- resp
				return
			}

			// errors other than EOF pose a problem
			if err != io.EOF {
				_, resp.Error = "", err
				ch <- resp
				return
			}

			// check if file has reduced in size which could mean file rotation or someone playing with it
			fileStat, err := file.Stat()
			if err != nil {
				_, resp.Error = "", err
				ch <- resp
				return
			}
			// get current file read offset
			currentRead, err := file.Seek(0, io.SeekCurrent)
			if err != nil {
				_, resp.Error = "", err
				ch <- resp
				return
			}
			// now if current read offset is greater than total file size,
			// then start reading from current file size
			if currentRead > fileStat.Size() {
				file.Seek(fileStat.Size(), io.SeekStart)
			}
		}

		if line != "" {
			// strip out new line char
			lineParts := strings.Split(line, "\n")
			matches := r.FindStringSubmatch(lineParts[0])

			if len(matches) == 2 {
				if fNum, err := strconv.ParseFloat(matches[1], 64); err == nil {
					dataFloats = append(dataFloats, fNum)
				}
			}

			if len(matches) > 2 {
				// for future implementation - plot multiple data stream on a single graph
				log.Fatal("Stopping: regex matched more than one group")
			}

			// ignore if line wasn't matched (i.e. length of matches less than 2)
		}
	}

	// set startPos to file content read so far (for next iteration of read)
	currentRead, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		_, resp.Error = "", err
		ch <- resp
		return
	}
	startPos = currentRead
	// return on channel
	resp.GraphId, resp.ConfigName, resp.Data, resp.FileName, resp.Position, resp.Error = graphId, configName, dataFloats, fileToRead, startPos, nil
	ch <- resp
}
