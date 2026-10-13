package signing

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"time"
)

type Identity struct {
	Key         *rsa.PrivateKey
	Certificate []byte
}

func CreateIdentity(path string) (Identity, error) {
	var id Identity
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		return id, e
	}
	serial, e := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if e != nil {
		return id, e
	}
	now := time.Now().UTC()
	tmpl := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: "koru application signing"}, NotBefore: now.Add(-time.Hour), NotAfter: now.AddDate(30, 0, 0), KeyUsage: x509.KeyUsageDigitalSignature}
	cert, e := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if e != nil {
		return id, e
	}
	b := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	b = append(b, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})...)
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return id, e
	}
	ok := false
	defer func() {
		f.Close()
		if !ok {
			os.Remove(path)
		}
	}()
	if _, e = f.Write(b); e != nil {
		return id, e
	}
	if e = f.Sync(); e != nil {
		return id, e
	}
	if e = f.Close(); e != nil {
		return id, e
	}
	ok = true
	return Identity{key, cert}, nil
}
func LoadIdentity(path string) (Identity, error) {
	var id Identity
	info, e := os.Lstat(path)
	if e != nil {
		return id, e
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return id, errors.New("signing identity must be a regular file with mode 0600")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return id, e
	}
	for len(b) > 0 {
		block, rest := pem.Decode(b)
		if block == nil {
			return id, errors.New("invalid PEM identity")
		}
		b = rest
		switch block.Type {
		case "RSA PRIVATE KEY":
			if id.Key != nil {
				return id, errors.New("duplicate private key")
			}
			id.Key, e = x509.ParsePKCS1PrivateKey(block.Bytes)
			if e != nil {
				return id, e
			}
		case "CERTIFICATE":
			if id.Certificate != nil {
				return id, errors.New("duplicate certificate")
			}
			id.Certificate = block.Bytes
		default:
			return id, errors.New("unsupported PEM block")
		}
	}
	if e = validateIdentity(id); e != nil {
		return Identity{}, e
	}
	return id, nil
}
func validateIdentity(id Identity) error {
	if id.Key == nil || len(id.Certificate) == 0 {
		return errors.New("missing signing key or certificate")
	}
	if id.Key.N.BitLen() < 2048 {
		return errors.New("RSA key must be at least 2048 bits")
	}
	if e := id.Key.Validate(); e != nil {
		return e
	}
	cert, e := x509.ParseCertificate(id.Certificate)
	if e != nil {
		return e
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok || pub.N.Cmp(id.Key.N) != 0 || pub.E != id.Key.E {
		return errors.New("certificate does not match signing key")
	}
	return nil
}
