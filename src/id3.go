// Copyright 2011 Andrew Scherkus
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package id3

import (
	"bufio"
	"fmt"
	"io"
)

type File struct {
	Header ID3v2Header

	Name   string
	Artist string
	Album  string
	Year   string
	Track  string
	Disc   string
	Genre  string
	Length string
}

type ID3v2Header struct {
	Version           int
	MinorVersion      int
	Unsynchronization bool
	Extended          bool
	Experimental      bool
	Footer            bool
	Size              int32
}

type ID3Parser interface {
	HasFrame() bool
	ReadFrame(file *File)
}

func Read(reader io.Reader) *File {
	file := new(File)
	bufReader := bufio.NewReader(reader)
	if !isID3Tag(bufReader) {
		return nil
	}

	parseID3v2Header(file, bufReader)
	limitReader := bufio.NewReader(io.LimitReader(bufReader, int64(file.Header.Size)))
	var parser ID3Parser
	if file.Header.Version == 2 {
		parser = NewID3v22Parser(limitReader)
	} else if file.Header.Version == 3 {
		parser = NewID3v23Parser(limitReader)
	} else if file.Header.Version == 4 {
		parser = NewID3v24Parser(limitReader)
	} else {
		panic(fmt.Sprintf("Unrecognized ID3v2 version: %d", file.Header.Version))
	}

	for parser.HasFrame() {
		parser.ReadFrame(file)
	}

	return file
}

func isID3Tag(reader *bufio.Reader) bool {
	data, err := reader.Peek(3)
	if len(data) < 3 || err != nil {
		return false
	}
	return data[0] == 'I' && data[1] == 'D' && data[2] == '3'
}

func parseID3v2Header(file *File, reader io.Reader) {
	data := make([]byte, 10)
	reader.Read(data)
	file.Header.Version = int(data[3])
	file.Header.MinorVersion = int(data[4])
	file.Header.Unsynchronization = data[5]&1<<7 != 0
	file.Header.Extended = data[5]&1<<6 != 0
	file.Header.Experimental = data[5]&1<<5 != 0
	file.Header.Footer = data[5]&1<<4 != 0
	file.Header.Size = parseSize(data[6:])
}
