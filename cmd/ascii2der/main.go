// Copyright 2015 The DER ASCII Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/der-ascii/ascii2der"
)

// pairs conforms to flag.Value. Each time Set() is called, it collects another
// k=v pair into itself.
type pairs map[string]string

func (p pairs) String() string {
	return ""
}

func (p pairs) Set(pair string) error {
	if pair == "" || p == nil {
		return nil
	}

	split := strings.SplitN(pair, "=", 2)
	if len(split) != 2 {
		return fmt.Errorf("missing \"=\": %q", pair)
	}

	p[split[0]] = split[1]
	return nil
}

var defines = make(map[string]string)
var fileDefines = make(map[string]string)

func init() {
	flag.Var(pairs(defines), "d",
		`pair of the form a=b; define("a", "b") is inserted at the start of the input`+
			"\nmay occur multiple times")
	flag.Var(pairs(fileDefines), "df",
		`like -d, except the second value is interpreted as a binary file to read`+
			"\nmay occur multiple times")
}

var inPath = flag.String("i", "", "input file to use (defaults to stdin)")
var outPath = flag.String("o", "", "output file to use (defaults to stdout)")
var pemType = flag.String("pem", "", "if provided, format the output as a PEM block with this type")

func readAll(path string) []byte {
	var file *os.File
	if path == "" {
		file = os.Stdin
	} else {
		var err error
		file, err = os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %s\n", path, err)
			os.Exit(1)
		}
		defer file.Close()
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %s\n", path, err)
		os.Exit(1)
	}

	return buf
}

func main() {
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION...]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	inBytes := readAll(*inPath)
	scanner := ascii2der.NewScanner(string(inBytes))
	scanner.SetFile(*inPath)

	scanner.Vars = make(map[string][]byte)
	for k, v := range defines {
		if _, ok := scanner.Vars[k]; ok {
			fmt.Fprintf(os.Stderr, "Error: tried to define %q with flags twice\n", k)
			os.Exit(1)
		}
		scanner.Vars[k] = []byte(v)
	}
	for k, v := range fileDefines {
		if _, ok := scanner.Vars[k]; ok {
			fmt.Fprintf(os.Stderr, "Error: tried to define %q with flags twice\n", k)
			os.Exit(1)
		}
		scanner.Vars[k] = readAll(v)
	}

	outBytes, err := scanner.Exec()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Syntax error: %s\n", err)
		os.Exit(1)
	}

	if *pemType != "" {
		outBytes = pem.EncodeToMemory(&pem.Block{
			Type:  *pemType,
			Bytes: outBytes,
		})
	}

	outFile := os.Stdout
	if *outPath != "" {
		var err error
		outFile, err = os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %s\n", *outPath, err)
			os.Exit(1)
		}
		defer outFile.Close()
	}
	_, err = outFile.Write(outBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %s\n", err)
		os.Exit(1)
	}
}
