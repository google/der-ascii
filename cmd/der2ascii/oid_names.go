// Copyright 2016 The DER ASCII Authors. All Rights Reserved.
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

// This file is generated by make_oid_names.go. Do not edit by hand.
// To regenerate, run "go run util/make_oid_names.go" from the top-level directory.

package main

var oidNames = []struct {
	oid  []byte
	name string
}{
	{[]byte{0x2b, 0x81, 0x4, 0x0, 0x21}, "secp224r1"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x3, 0x1, 0x7}, "secp256r1"},
	{[]byte{0x2b, 0x81, 0x4, 0x0, 0x22}, "secp384r1"},
	{[]byte{0x2b, 0x81, 0x4, 0x0, 0x23}, "secp521r1"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x1, 0x1}, "prime-field"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x1, 0x2}, "characteristic-two-field"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x1, 0x2, 0x3, 0x1}, "gnBasis"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x1, 0x2, 0x3, 0x2}, "tpBasis"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x1, 0x2, 0x3, 0x3}, "ppBasis"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x2, 0x2}, "md2"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x2, 0x4}, "md4"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x2, 0x5}, "md5"},
	{[]byte{0x2b, 0xe, 0x3, 0x2, 0x1a}, "sha1"},
	{[]byte{0x60, 0x86, 0x48, 0x1, 0x65, 0x3, 0x4, 0x2, 0x4}, "sha224"},
	{[]byte{0x60, 0x86, 0x48, 0x1, 0x65, 0x3, 0x4, 0x2, 0x1}, "sha256"},
	{[]byte{0x60, 0x86, 0x48, 0x1, 0x65, 0x3, 0x4, 0x2, 0x2}, "sha384"},
	{[]byte{0x60, 0x86, 0x48, 0x1, 0x65, 0x3, 0x4, 0x2, 0x3}, "sha512"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0x8}, "mgf1"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0x1}, "rsaEncryption"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0xa}, "rsassa-pss"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x2, 0x1}, "ecPublicKey"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x38, 0x4, 0x1}, "dsa"},
	{[]byte{0x2b, 0x65, 0x6e}, "x25519"},
	{[]byte{0x2b, 0x65, 0x6f}, "x448"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0x2}, "md2WithRSAEncryption"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0x3}, "md4WithRSAEncryption"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0x4}, "md5WithRSAEncryption"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0x5}, "sha1WithRSAEncryption"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0xe}, "sha224WithRSAEncryption"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0xb}, "sha256WithRSAEncryption"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0xc}, "sha384WithRSAEncryption"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x1, 0xd}, "sha512WithRSAEncryption"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x38, 0x4, 0x3}, "dsa-with-sha1"},
	{[]byte{0x60, 0x86, 0x48, 0x1, 0x65, 0x3, 0x4, 0x3, 0x1}, "dsa-with-sha224"},
	{[]byte{0x60, 0x86, 0x48, 0x1, 0x65, 0x3, 0x4, 0x3, 0x2}, "dsa-with-sha256"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x4, 0x1}, "ecdsa-with-SHA1"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x4, 0x3, 0x1}, "ecdsa-with-SHA224"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x4, 0x3, 0x2}, "ecdsa-with-SHA256"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x4, 0x3, 0x3}, "ecdsa-with-SHA384"},
	{[]byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x4, 0x3, 0x4}, "ecdsa-with-SHA512"},
	{[]byte{0x2b, 0x65, 0x70}, "ed25519"},
	{[]byte{0x2b, 0x65, 0x71}, "ed448"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x1, 0x1}, "authorityInfoAccess"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x1, 0x7}, "ipAddrBlocks"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x1, 0x8}, "autonomousSysIds"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x1, 0xb}, "subjectInfoAccess"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x1, 0x1c}, "ipAddrBlocks-v2"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x1, 0x1d}, "autonomousSysIds-v2"},
	{[]byte{0x55, 0x1d, 0x9}, "subjectDirectoryAttributes"},
	{[]byte{0x55, 0x1d, 0xe}, "subjectKeyIdentifier"},
	{[]byte{0x55, 0x1d, 0xf}, "keyUsage"},
	{[]byte{0x55, 0x1d, 0x10}, "privateKeyUsagePeriod"},
	{[]byte{0x55, 0x1d, 0x11}, "subjectAltName"},
	{[]byte{0x55, 0x1d, 0x12}, "issuerAltName"},
	{[]byte{0x55, 0x1d, 0x13}, "basicConstraints"},
	{[]byte{0x55, 0x1d, 0x14}, "cRLNumber"},
	{[]byte{0x55, 0x1d, 0x15}, "reasonCode"},
	{[]byte{0x55, 0x1d, 0x17}, "instructionCode"},
	{[]byte{0x55, 0x1d, 0x18}, "invalidityDate"},
	{[]byte{0x55, 0x1d, 0x1b}, "deltaCRLIndicator"},
	{[]byte{0x55, 0x1d, 0x1c}, "issuingDistributionPoint"},
	{[]byte{0x55, 0x1d, 0x1d}, "certificateIssuer"},
	{[]byte{0x55, 0x1d, 0x1e}, "nameConstraints"},
	{[]byte{0x55, 0x1d, 0x1f}, "cRLDistributionPoints"},
	{[]byte{0x55, 0x1d, 0x20}, "certificatePolicies"},
	{[]byte{0x55, 0x1d, 0x21}, "policyMappings"},
	{[]byte{0x55, 0x1d, 0x23}, "authorityKeyIdentifier"},
	{[]byte{0x55, 0x1d, 0x24}, "policyConstraints"},
	{[]byte{0x55, 0x1d, 0x25}, "extKeyUsage"},
	{[]byte{0x55, 0x1d, 0x2e}, "freshestCRL"},
	{[]byte{0x55, 0x1d, 0x36}, "inhibitAnyPolicy"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x3, 0x1}, "serverAuth"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x3, 0x2}, "clientAuth"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x3, 0x3}, "codeSigning"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x3, 0x4}, "emailProtection"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x3, 0x8}, "timeStamping"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x3, 0x9}, "OCSPSigning"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x3, 0x1e}, "bgpsec-router"},
	{[]byte{0x55, 0x1d, 0x25, 0x0}, "anyExtendedKeyUsage"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x2, 0x2}, "unotice"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0xe, 0x2}, "ipAddr-asNumber"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0xe, 0x3}, "ipAddr-asNumber-v2"},
	{[]byte{0x55, 0x1d, 0x20, 0x0}, "anyPolicy"},
	{[]byte{0x67, 0x81, 0xc, 0x1, 0x2, 0x1}, "domain-validated"},
	{[]byte{0x67, 0x81, 0xc, 0x1, 0x2, 0x2}, "organization-validated"},
	{[]byte{0x67, 0x81, 0xc, 0x1, 0x2, 0x3}, "individual-validated"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x30, 0x1}, "ocsp"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x30, 0x2}, "caIssuers"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x30, 0x5}, "caRepository"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x30, 0xa}, "rpkiManifest"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x30, 0xb}, "signedObject"},
	{[]byte{0x2b, 0x6, 0x1, 0x5, 0x5, 0x7, 0x30, 0xd}, "rpkiNotify"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x1}, "emailAddress"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x2}, "unstructuredName"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x3}, "contentType"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x4}, "messageDigest"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x5}, "signingTime"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x6}, "counterSignature"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x7}, "challengePassword"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x8}, "unstructuredAddress"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x9}, "extendedCertificateAttributes"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0xa}, "issuerAndSerialNumber"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0xb}, "passwordCheck"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0xc}, "publicKey"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0xd}, "signingDescription"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0xe}, "extensionRequest"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0xf}, "smimeCapabilities"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x14}, "friendlyName"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x15}, "localKeyId"},
	{[]byte{0x55, 0x4, 0x3}, "commonName"},
	{[]byte{0x55, 0x4, 0x5}, "serialNumber"},
	{[]byte{0x55, 0x4, 0x6}, "countryName"},
	{[]byte{0x55, 0x4, 0x7}, "localityName"},
	{[]byte{0x55, 0x4, 0x8}, "stateOrProvinceName"},
	{[]byte{0x55, 0x4, 0x9}, "streetAddress"},
	{[]byte{0x55, 0x4, 0xa}, "organizationName"},
	{[]byte{0x55, 0x4, 0xb}, "organizationUnitName"},
	{[]byte{0x55, 0x4, 0xc}, "title"},
	{[]byte{0x55, 0x4, 0x11}, "postalCode"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0x1}, "receiptRequest"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0x2}, "securityLabel"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0x3}, "mlExpandHistory"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0x4}, "contentHint"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0x5}, "msgSigDigest"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0x7}, "contentIdentifier"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0x9}, "equivalentLabels"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0xa}, "contentReference"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0xb}, "encrypKeyPref"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x2, 0xc}, "signingCertificate"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0xb, 0x1}, "preferBinaryInside"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x7, 0x1}, "data"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x7, 0x2}, "signedData"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x7, 0x3}, "envelopedData"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x7, 0x4}, "signedAndEnvelopedData"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x7, 0x5}, "digestedData"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x7, 0x6}, "encryptedData"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x1}, "receipt"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x2}, "authData"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x6}, "contentInfo"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x18}, "routeOriginAuthz"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x1a}, "rpkiManifest"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x23}, "rpkiGhostbusters"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x2f}, "geofeedCSVwithCRLF"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x30}, "signedChecklist"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x31}, "ASPA"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x9, 0x10, 0x1, 0x32}, "signedTAL"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0xa, 0x1, 0x1}, "keyBag"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0xa, 0x1, 0x2}, "pkcs-8ShroudedKeyBag"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0xa, 0x1, 0x3}, "certBag"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0xa, 0x1, 0x4}, "crlBag"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0xa, 0x1, 0x5}, "secretBag"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0xa, 0x1, 0x6}, "safeContentsBag"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0x1, 0x1}, "pbeWithSHAAnd128BitRC4"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0x1, 0x2}, "pbeWithSHAAnd40BitRC4"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0x1, 0x3}, "pbeWithSHAAnd3-KeyTripleDES-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0x1, 0x4}, "pbeWithSHAAnd2-KeyTripleDES-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0x1, 0x5}, "pbeWithSHAAnd128BitRC2-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0xc, 0x1, 0x6}, "pbewithSHAAnd40BitRC2-CBC"},
	{[]byte{0x2b, 0x6, 0x1, 0x4, 0x1, 0xd6, 0x79, 0x2, 0x4, 0x2}, "embeddedSCTList"},
	{[]byte{0x2b, 0x6, 0x1, 0x4, 0x1, 0xd6, 0x79, 0x2, 0x4, 0x3}, "ctPoison"},
	{[]byte{0x2b, 0x6, 0x1, 0x4, 0x1, 0xd6, 0x79, 0x2, 0x4, 0x4}, "ctPrecertificateSigning"},
	{[]byte{0x2b, 0x6, 0x1, 0x4, 0x1, 0xd6, 0x79, 0x2, 0x4, 0x5}, "ocspSCTList"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x5, 0x1}, "pbeWithMD2AndDES-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x5, 0x3}, "pbeWithMD5AndDES-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x5, 0x4}, "pbeWithMD2AndRC2-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x5, 0x6}, "pbeWithMD5AndRC2-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x5, 0xa}, "pbeWithSHA1AndDES-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x5, 0xb}, "pbeWithSHA1AndRC2-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x5, 0xc}, "PBKDF2"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x5, 0xd}, "PBES2"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x1, 0x5, 0xe}, "PBMAC1"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x2, 0x7}, "hmacWithSHA1"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x2, 0x8}, "hmacWithSHA224"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x2, 0x9}, "hmacWithSHA256"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x2, 0xa}, "hmacWithSHA384"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x2, 0xb}, "hmacWithSHA512"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x3, 0x2}, "RC2-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x3, 0x7}, "DES-EDE3-CBC"},
	{[]byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0xd, 0x3, 0x9}, "RC5-CBC-Pad"},
	{[]byte{0x60, 0x86, 0x48, 0x1, 0x65, 0x3, 0x4, 0x1, 0x2}, "AES-128-CBC"},
	{[]byte{0x60, 0x86, 0x48, 0x1, 0x65, 0x3, 0x4, 0x1, 0x16}, "AES-192-CBC"},
	{[]byte{0x60, 0x86, 0x48, 0x1, 0x65, 0x3, 0x4, 0x1, 0x2a}, "AES-256-CBC"},
}
