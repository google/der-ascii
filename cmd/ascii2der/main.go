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
)

var inPath = flag.String("i", "", "input file to use (defaults to stdin)")
var outPath = flag.String("o", "", "output file to use (defaults to stdout)")
var pemType = flag.String("pem", "", "if provided, format the output as a PEM block with this type")

func main() {
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION...]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	inFile := os.Stdin
	if *inPath != "" {
		var err error
		inFile, err = os.Open(*inPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %s\n", *inPath, err)
			os.Exit(1)
		}
		defer inFile.Close()
	}

	inBytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %s\n", err)
		os.Exit(1)
	}

	outBytes, err := asciiToDER(string(inBytes))
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
