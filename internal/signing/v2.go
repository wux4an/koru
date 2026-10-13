// Package signing implements APK Signature Scheme v2 using the Go standard library.
// https://source.android.com/docs/security/features/apksigning/v2
package signing

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
)

const magic = "APK Sig Block 42"
const v2ID uint32 = 0x7109871a
const algorithm uint32 = 0x0103

func u32(v uint32) []byte { b := make([]byte, 4); binary.LittleEndian.PutUint32(b, v); return b }
func u64(v uint64) []byte { b := make([]byte, 8); binary.LittleEndian.PutUint64(b, v); return b }
func lp(b []byte) []byte  { return append(u32(uint32(len(b))), b...) }

// Layout rejects ZIP64, multipart archives and ambiguous/trailing EOCD records.
func Layout(b []byte) (central, eocd int, err error) {
	for p := len(b) - 22; p >= 0 && p >= len(b)-65557; p-- {
		if binary.LittleEndian.Uint32(b[p:]) != 0x06054b50 {
			continue
		}
		if p+22+int(binary.LittleEndian.Uint16(b[p+20:])) != len(b) {
			continue
		}
		if binary.LittleEndian.Uint16(b[p+4:]) != 0 || binary.LittleEndian.Uint16(b[p+6:]) != 0 || binary.LittleEndian.Uint16(b[p+8:]) != binary.LittleEndian.Uint16(b[p+10:]) {
			return 0, 0, errors.New("multipart ZIP unsupported")
		}
		size, off := binary.LittleEndian.Uint32(b[p+12:]), binary.LittleEndian.Uint32(b[p+16:])
		if size == 0xffffffff || off == 0xffffffff || binary.LittleEndian.Uint16(b[p+10:]) == 65535 {
			return 0, 0, errors.New("ZIP64 unsupported")
		}
		if uint64(off)+uint64(size) != uint64(p) {
			return 0, 0, errors.New("central directory must end at EOCD")
		}
		return int(off), p, nil
	}
	return 0, 0, errors.New("valid ZIP EOCD not found")
}
func digest(parts ...[]byte) []byte {
	var hashes []byte
	var count uint32
	for _, part := range parts {
		for len(part) > 0 {
			n := len(part)
			if n > 1<<20 {
				n = 1 << 20
			}
			h := sha256.New()
			h.Write([]byte{0xa5})
			h.Write(u32(uint32(n)))
			h.Write(part[:n])
			hashes = append(hashes, h.Sum(nil)...)
			count++
			part = part[n:]
		}
	}
	h := sha256.New()
	h.Write([]byte{0x5a})
	h.Write(u32(count))
	h.Write(hashes)
	return h.Sum(nil)
}
func SignV2(b []byte, id Identity) ([]byte, error) {
	if e := validateIdentity(id); e != nil {
		return nil, e
	}
	cd, end, e := Layout(b)
	if e != nil {
		return nil, e
	}
	if cd >= 16 && string(b[cd-16:cd]) == magic {
		return nil, errors.New("APK already contains a signing block")
	}
	d := digest(b[:cd], b[cd:end], b[end:])
	dig := lp(append(u32(algorithm), lp(d)...))
	signed := append(lp(dig), lp(lp(id.Certificate))...)
	signed = append(signed, lp(nil)...)
	h := sha256.Sum256(signed)
	sig, e := rsa.SignPKCS1v15(rand.Reader, id.Key, crypto.SHA256, h[:])
	if e != nil {
		return nil, e
	}
	pub, e := x509.MarshalPKIXPublicKey(&id.Key.PublicKey)
	if e != nil {
		return nil, e
	}
	signer := append(lp(signed), lp(lp(append(u32(algorithm), lp(sig)...)))...)
	signer = append(signer, lp(pub)...)
	value := lp(lp(signer))
	pair := append(u64(uint64(4+len(value))), u32(v2ID)...)
	pair = append(pair, value...)
	size := uint64(len(pair) + 24)
	block := append(u64(size), pair...)
	block = append(block, u64(size)...)
	block = append(block, []byte(magic)...)
	if uint64(cd)+uint64(len(block)) > 0xffffffff {
		return nil, errors.New("APK exceeds ZIP32 offsets")
	}
	result := append([]byte(nil), b[:cd]...)
	result = append(result, block...)
	result = append(result, b[cd:]...)
	binary.LittleEndian.PutUint32(result[end+len(block)+16:], uint32(cd+len(block)))
	return result, nil
}

type cursor struct {
	b   []byte
	err error
}

func (c *cursor) take(n uint64) []byte {
	if c.err != nil {
		return nil
	}
	if n > uint64(len(c.b)) {
		c.err = errors.New("truncated signing block")
		return nil
	}
	v := c.b[:n]
	c.b = c.b[n:]
	return v
}
func (c *cursor) uint32() uint32 {
	b := c.take(4)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}
func (c *cursor) length() []byte { return c.take(uint64(c.uint32())) }
func (c *cursor) done() error {
	if c.err != nil {
		return c.err
	}
	if len(c.b) != 0 {
		return errors.New("unexpected signing data")
	}
	return nil
}
func VerifyV2(b []byte) ([]byte, error) {
	cd, end, e := Layout(b)
	if e != nil {
		return nil, e
	}
	if cd < 32 || string(b[cd-16:cd]) != magic {
		return nil, errors.New("APK v2 signing block not found")
	}
	size := binary.LittleEndian.Uint64(b[cd-24:])
	if size < 24 || size > uint64(cd)-8 {
		return nil, errors.New("invalid signing block size")
	}
	start := cd - int(size) - 8
	if binary.LittleEndian.Uint64(b[start:]) != size {
		return nil, errors.New("signing block size mismatch")
	}
	pairs := cursor{b: b[start+8 : cd-24]}
	var value []byte
	for len(pairs.b) > 0 && pairs.err == nil {
		n := pairs.take(8)
		if n == nil {
			break
		}
		length := binary.LittleEndian.Uint64(n)
		if length < 4 {
			return nil, errors.New("invalid signing pair")
		}
		pair := pairs.take(length)
		if pair == nil {
			break
		}
		id := binary.LittleEndian.Uint32(pair)
		if id == v2ID {
			if value != nil {
				return nil, errors.New("duplicate v2 signing pair")
			}
			value = pair[4:]
		}
	}
	if e = pairs.done(); e != nil {
		return nil, e
	}
	if value == nil {
		return nil, errors.New("v2 signer missing")
	}
	outer := cursor{b: value}
	signers := cursor{b: outer.length()}
	if e = outer.done(); e != nil {
		return nil, e
	}
	var certificate []byte
	count := 0
	for len(signers.b) > 0 && signers.err == nil {
		signer := cursor{b: signers.length()}
		signed := signer.length()
		signatures := cursor{b: signer.length()}
		public := signer.length()
		if e = signer.done(); e != nil {
			return nil, e
		}
		pub, e := x509.ParsePKIXPublicKey(public)
		if e != nil {
			return nil, e
		}
		key, ok := pub.(*rsa.PublicKey)
		if !ok || key.N.BitLen() < 2048 || key.N.BitLen() > 16384 {
			return nil, errors.New("unsupported signing key")
		}
		signature := cursor{b: signatures.length()}
		alg := signature.uint32()
		sig := signature.length()
		if e = signature.done(); e != nil {
			return nil, e
		}
		if e = signatures.done(); e != nil {
			return nil, e
		}
		if alg != algorithm {
			return nil, errors.New("unsupported signature algorithm")
		}
		h := sha256.Sum256(signed)
		if e = rsa.VerifyPKCS1v15(key, crypto.SHA256, h[:], sig); e != nil {
			return nil, fmt.Errorf("APK signature: %w", e)
		}
		data := cursor{b: signed}
		digests := cursor{b: data.length()}
		certs := cursor{b: data.length()}
		attributes := data.length()
		if len(attributes) != 0 {
			return nil, errors.New("unsupported signer attributes")
		}
		if e = data.done(); e != nil {
			return nil, e
		}
		dg := cursor{b: digests.length()}
		if dg.uint32() != alg {
			return nil, errors.New("digest/signature algorithm mismatch")
		}
		expected := dg.length()
		if e = dg.done(); e != nil {
			return nil, e
		}
		if e = digests.done(); e != nil {
			return nil, e
		}
		cert := certs.length()
		parsed, e := x509.ParseCertificate(cert)
		if e != nil {
			return nil, e
		}
		for len(certs.b) > 0 && certs.err == nil {
			if _, e = x509.ParseCertificate(certs.length()); e != nil {
				return nil, e
			}
		}
		if e = certs.done(); e != nil {
			return nil, e
		}
		if !bytes.Equal(parsed.RawSubjectPublicKeyInfo, public) {
			return nil, errors.New("certificate/public key mismatch")
		}
		eo := append([]byte(nil), b[end:]...)
		binary.LittleEndian.PutUint32(eo[16:], uint32(start))
		actual := digest(b[:start], b[cd:end], eo)
		if !bytes.Equal(expected, actual) {
			return nil, errors.New("APK content digest mismatch")
		}
		if count == 0 {
			certificate = append([]byte(nil), cert...)
		}
		count++
	}
	if e = signers.done(); e != nil {
		return nil, e
	}
	if count == 0 {
		return nil, errors.New("APK contains no signers")
	}
	return certificate, nil
}
