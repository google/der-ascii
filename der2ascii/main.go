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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var outPath = flag.String("o", "", "output file to use (defaults to stdout)")

func main() {
	flag.Parse()

	var err error
	inFile := os.Stdin
	if flag.NArg() > 0 {
		if flag.NArg() == 1 {
			inFile, err = os.Open(flag.Arg(0))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening %s: %s\n", flag.Arg(0), err)
				os.Exit(1)
			}
			defer inFile.Close()
		} else {
			fmt.Fprintf(os.Stderr, "Usage: %s [INPUT] [-o OUTPUT]\n", os.Args[0])
			os.Exit(1)
		}
	}

	inBytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %s\n", err)
		os.Exit(1)
	}

	outFile := os.Stdout
	if *outPath != "" {
		outFile, err = os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %s\n", *outPath, err)
			os.Exit(1)
		}
		defer outFile.Close()
	}
	_, err = outFile.Write([]byte(derToASCII(inBytes)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %s\n", err)
		os.Exit(1)
	}
}
