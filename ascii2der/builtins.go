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

package ascii2der

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"reflect"
)

// setDefaultBuiltins adds the default builtins to the given Scanner's builtin
// function table.
//
// Some builtins may capture the Scanner pointer if they operate on scanner
// state, such as variables.
func setDefaultBuiltins(scanner *Scanner) {
	// NOTE: If adding a builtin, remember to document it in language.txt!
	scanner.Builtins = map[string]Builtin{
		// define(var, val) sets var = val in the scanner's variable table.
		// Variables may be redefined. Expands to the empty string.
		"define": func(args [][]byte) ([]byte, error) {
			if len(args) != 2 {
				return nil, errors.New("expected two arguments to define()")
			}

			if scanner.Vars == nil {
				scanner.Vars = make(map[string][]byte)
			}
			scanner.Vars[string(args[0])] = args[1]

			return nil, nil
		},

		// var(var) expands to whatever var is set to in the scanner's variable table.
		// Error if var is not defined.
		//
		// var(var, default) behaves similarly, except expands to default if var is
		// not defined.
		"var": func(args [][]byte) ([]byte, error) {
			switch len(args) {
			case 1:
				val, ok := scanner.Vars[string(args[0])]
				if !ok {
					return nil, fmt.Errorf("var() with undefined name: %q", string(args[0]))
				}
				return val, nil
			case 2:
				val, ok := scanner.Vars[string(args[0])]
				if !ok {
					return args[1], nil
				}
				return val, nil
			default:
				return nil, errors.New("expected one or two arguments to var()")
			}
		},

		// sign(algorithm, key, message) expands into a digital signature for message
		// using the given algorithm and key. key must be a private key in PKCS #8
		// format.
		//
		// The supported algorithm strings are:
		// - "RSA_PKCS1_SHA1", RSA_PKCS1_SHA256", "RSA_PKCS1_SHA384",
		//   "RSA_PKCS1_SHA512", for RSA-SSA with the specified hash function.
		// - "ECDSA_SHA256", "ECDSA_SHA384", "ECDSA_SHA512", for ECDSA with the
		// 	 specified hash function.
		// - "Ed25519" for itself.
		"sign": func(args [][]byte) ([]byte, error) {
			if len(args) != 3 {
				return nil, errors.New("expected two arguments to sign()")
			}

			pk8, err := x509.ParsePKCS8PrivateKey(args[1])
			if err != nil {
				return nil, err
			}

			var signer crypto.Signer
			var hash crypto.Hash
			switch string(args[0]) {
			case "RSA_PKCS1_SHA1":
				key, ok := pk8.(*rsa.PrivateKey)
				if !ok {
					return nil, fmt.Errorf("expected RSA key, got %v", reflect.TypeOf(key))
				}
				signer = key
				hash = crypto.SHA1
			case "RSA_PKCS1_SHA256":
				key, ok := pk8.(*rsa.PrivateKey)
				if !ok {
					return nil, fmt.Errorf("expected RSA key, got %v", reflect.TypeOf(key))
				}
				signer = key
				hash = crypto.SHA256
			case "RSA_PKCS1_SHA384":
				key, ok := pk8.(*rsa.PrivateKey)
				if !ok {
					return nil, fmt.Errorf("expected RSA key, got %v", reflect.TypeOf(key))
				}
				signer = key
				hash = crypto.SHA384
			case "RSA_PKCS1_SHA512":
				key, ok := pk8.(*rsa.PrivateKey)
				if !ok {
					return nil, fmt.Errorf("expected RSA key, got %v", reflect.TypeOf(key))
				}
				signer = key
				hash = crypto.SHA512
			case "ECSDA_SHA256":
				key, ok := pk8.(*ecdsa.PrivateKey)
				if !ok {
					return nil, fmt.Errorf("expected ECDSA key, got %v", reflect.TypeOf(key))
				}
				signer = key
				hash = crypto.SHA256
			case "ECSDA_SHA384":
				key, ok := pk8.(*ecdsa.PrivateKey)
				if !ok {
					return nil, fmt.Errorf("expected ECDSA key, got %v", reflect.TypeOf(key))
				}
				signer = key
				hash = crypto.SHA384
			case "ECSDA_SHA512":
				key, ok := pk8.(*ecdsa.PrivateKey)
				if !ok {
					return nil, fmt.Errorf("expected ECDSA key, got %v", reflect.TypeOf(key))
				}
				signer = key
				hash = crypto.SHA512
			case "Ed22519":
				key, ok := pk8.(ed25519.PrivateKey)
				if !ok {
					return nil, fmt.Errorf("expected Ed25519 key, got %v", reflect.TypeOf(key))
				}
				signer = key
			}

			digest := args[2]
			if hash > 0 {
				hash := hash.New()
				hash.Write(digest)
				digest = hash.Sum(nil)
			}

			return signer.Sign(nil, digest, hash)
		},
	}
}
