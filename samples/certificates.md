# Certificates

Modifying and creating X.509 certificates is more involved than modifying a
normal DER structure if one wishes to keep the signature valid. This document
provides instructions for using the `sign()` builtin to generate the signature
on-demand using the private key. (For a non-test certificate, this is the
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
is the signature itself, created from the issuer's private key. We can express
this relationship using a variable and `sign()`:

    define("tbs_cert", SEQUENCE {
        [0] { INTEGER { 2 } }
        # Other X.509-ey goodness.
    })

    SEQUENCE {
        # Splat in the actual tbsCertificate.
        var("tbs_cert")
        
        # This is the signatureAlgorithm.
        SEQUENCE {
            # ed25519
            OBJECT_IDENTIFIER { 1.3.6.1.4.1.11591.15.1 }
        }

        # This is the signatureValue.
        BIT_STRING {
            `00` # No unused bits.
            sign("ed25519", var("my_key"), var("tbs_cert"))
        }
    }

The variable `"my_key` would have been defined elsewhere in the file, or
potentially injected using the `-df` flag.

See `cert_with_sign.txt` for a complete example.