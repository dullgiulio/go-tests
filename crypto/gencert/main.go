package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"
)

func hash(b []byte) ([]byte, error) {
	h := sha1.New()
	_, err := h.Write(b)
	if err != nil {
		return nil, err
	}
	c := h.Sum(nil)
	return c, nil
}

func encBase64(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

// TODO: represent which curve is used!
func ecdsaBytes(key interface{}) []byte {
	if k, ok := key.(*ecdsa.PrivateKey); ok {
		return k.D.Bytes()
	}
	if k, ok := key.(*ecdsa.PublicKey); ok {
		x := k.X.Bytes()
		y := k.Y.Bytes()
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "%d\n", len(x))
		buf.Write(x)
		buf.Write(y)
		return buf.Bytes()
	}
	panic("key used is not ECDSA")
}

func bytesEcdsa(c elliptic.Curve, privbs, pubbs []byte) (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	priv := &ecdsa.PrivateKey{}
	priv.Curve = c
	priv.D = &big.Int{}
	priv.D.SetBytes(privbs)
	pub := &ecdsa.PublicKey{}
	pub.Curve = c
	nl := bytes.IndexByte(pubbs, '\n')
	if nl < 0 {
		return nil, nil, errors.New("public ECDSA key contains no length for X")
	}
	xlen, err := strconv.ParseInt(string(pubbs[0:nl]), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("public ECDSA key length for X is not a number: %v", err)
	}
	xbs := pubbs[nl+1 : xlen]
	pubbs = pubbs[xlen:]
	pub.X = &big.Int{}
	pub.X.SetBytes(xbs)
	pub.Y = &big.Int{}
	pub.Y.SetBytes(pubbs)
	return priv, pub, nil
}

func main() {
	commonName := "*.t3env.int.kn" // TODO: configurable
	curve := elliptic.P224()       // Chosen at random, read which one to use
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Fatalf("cannot generate ECDSA key: %v", err)
	}
	now := time.Now()
	pub := priv.Public()
	authorityKeyId, err := hash(priv.D.Bytes()) // TODO: maybe use public key
	if err != nil {
		log.Fatalf("cannot compute checksum of public key: %v", err)
	}
	serialNumber := big.NewInt(int64(now.Nanosecond())) // Insecure, but who cares?
	usagesCA := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	name := pkix.Name{
		Country:            []string{"Estonia"},
		Organization:       []string{"TestOrg"},
		OrganizationalUnit: []string{"TestUnit"},
		Locality:           []string{"Locality"},
		Province:           []string{"Province"},
		StreetAddress:      []string{"Address"},
		PostalCode:         []string{"Postal Code"},
		SerialNumber:       fmt.Sprintf("%s", serialNumber),
		CommonName:         commonName,
	}
	tmpl := &x509.Certificate{
		IsCA:                  true,
		MaxPathLenZero:        true, // For CA
		KeyUsage:              usagesCA,
		DNSNames:              []string{commonName},               // For SAN
		NotAfter:              now.Add(10 * 365 * 24 * time.Hour), // roughly 10 years
		NotBefore:             now.Add(-24 * time.Hour),
		SerialNumber:          serialNumber,
		SignatureAlgorithm:    x509.ECDSAWithSHA512,
		BasicConstraintsValid: true,
		AuthorityKeyId:        authorityKeyId,
		Subject:               name,
		Issuer:                name,
		// ExtKeyUsage: ExtKeyUsageAny,
	}
	cert, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		log.Fatalf("cannot generate self-signed certificate: %v", err)
	}

	b64 := base64.RawStdEncoding

	encCert := b64.EncodeToString(cert)
	encPub := b64.EncodeToString(ecdsaBytes(pub))
	encPriv := b64.EncodeToString(ecdsaBytes(priv))
	keysPrint(encPub, encPriv)

	fmt.Printf("CERT:\n")
	fmt.Printf("%s\n", encCert)

	decPub, err := b64.DecodeString(encPub)
	if err != nil {
		log.Fatalf("cannot decode public key: %v", err)
	}
	decPriv, err := b64.DecodeString(encPriv)
	if err != nil {
		log.Fatalf("cannot decode private key: %v", err)
	}
	priv1, pub1, err := bytesEcdsa(curve, decPriv, decPub)
	if err != nil {
		log.Fatalf("cannot decode keys: %v", err)
	}
	encPub = b64.EncodeToString(ecdsaBytes(pub1))
	encPriv = b64.EncodeToString(ecdsaBytes(priv1))
	keysPrint(encPub, encPriv)
}

func keysPrint(pub, priv string) {
	fmt.Printf("Public key: %s\n", pub)
	fmt.Printf("Private key: %s\n", priv)
}
