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

func ReadFile(fileToRead string, maxLines int, startPos int64, re string) ([]float64, int64, error) {
	var dataFloats []float64
	r := regexp.MustCompile(re)

	file, err := os.Open(fileToRead)
	if err != nil {
		return dataFloats, startPos, err
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
				return dataFloats, startPos, err
			}

			// other than EOF is a problem
			if err != io.EOF {
				return dataFloats, startPos, err
			}

			// check if file has reduced in size which could mean file rotation
			fileStat, err := file.Stat()
			if err != nil {
				return dataFloats, startPos, err
			}

			// get current file read offset
			currentRead, err := file.Seek(0, io.SeekCurrent)
			if err != nil {
				return dataFloats, startPos, err
			}

			// if current read offset is greater than total file size,
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
				// future implementation
				log.Fatal("Stopping: regex matched more than one group")
			}

			// ignore if line wasn't matched (length of matches less than 2)
		}
	}

	// set startPos to file content read so far
	currentRead, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return dataFloats, startPos, err
	}
	startPos = currentRead

	return dataFloats, startPos, nil
}
