# Certificates

Modifying and creating X.509 certificates is more involved than modifying a
normal DER structure if one wishes to keep the signature valid. This document
provides instructions for fixing up a modified test certificate's signature if
the issuer's private key is available. (For a non-test certificate, this is the
CA's private key and is presumably unavailable.)

X.509 certificates are specified in [RFC 5280](https://tools.ietf.org/html/rfc5280).
The basic top-level structure is:

    Certificate  ::=  SEQUENCE  {
         tbsCertificate       TBSCertificate,
         signatureAlgorithm   AlgorithmIdentifier,
         signatureValue       BIT STRING  }

The `tbsCertificate` is a large structure with the contents of the certificate.
This includes the subject, issuer, public key, etc. The `signatureAlgorithm`
specifies the signature algorithm and parameters. Finally, the `signatureValue`
is the signature itself, created from the issuer's private key. This is the
field that must be fixed once the `tbsCertificate` is modified.

The signature is computed over the serialized `tbsCertificate`, so, using a
text editor, copy the `tbsCertificate` value into its own file, `tbs-cert.txt`.
Now sign that with the issuing private key. If using OpenSSL's command-line
tool, here is a sample command:

    ascii2der -i tbs-cert.txt | openssl dgst -sha256 -sign issuer_key.pem | \
        xxd -p -c 9999 > signature.txt

For other options, replace `-sha256` with a different digest or pass `-sigopt`.
See [OpenSSL's documentation](https://www.openssl.org/docs/manmaster/apps/dgst.html)
for details. Note that, for a valid certificate, the signature parameters
should match the `signatureAlgorithm` field. If using different signing
parameters, update it and the copy in the `tbsCertificate`.

Finally, in a text editor, replace the signature with the new one. X.509
defines certificates as BIT STRINGs, but every signature algorithm uses byte
strings, so include a leading zero to specify that no bits should be removed
from the end:

    BIT_STRING {
      `00` # No unused bits.
      `INSERT SIGNATURE HERE`
    }

Finally, use `ascii2der` to convert the certificate to DER.
