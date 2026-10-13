package signing

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"math/big"
	"testing"
	"time"
)

func testKey(t *testing.T) Identity {
	t.Helper()
	k, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		t.Fatal(e)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Unix(0, 0), NotAfter: time.Unix(4102444800, 0)}
	c, e := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &k.PublicKey, k)
	if e != nil {
		t.Fatal(e)
	}
	return Identity{Key: k, Certificate: c}
}
func unsigned(t *testing.T) []byte {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	f, e := z.Create("assets/www/index.html")
	if e != nil {
		t.Fatal(e)
	}
	f.Write([]byte("hello"))
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func TestSignVerifyAndTampering(t *testing.T) {
	key := testKey(t)
	b := unsigned(t)
	s, e := SignV2(b, key)
	if e != nil {
		t.Fatal(e)
	}
	cert, e := VerifyV2(s)
	if e != nil || !bytes.Equal(cert, key.Certificate) {
		t.Fatal(e)
	}
	s2, e := SignV2(b, key)
	if e != nil || !bytes.Equal(s, s2) {
		t.Fatal("not deterministic", e)
	}
	for _, p := range []int{0, 40, len(s) / 2, len(s) - 8} {
		bad := append([]byte(nil), s...)
		bad[p] ^= 1
		if _, e = VerifyV2(bad); e == nil {
			t.Fatal("tampering accepted", p)
		}
	}
}
func TestTruncatedSignatureRejected(t *testing.T) {
	s, e := SignV2(unsigned(t), testKey(t))
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < len(s); i++ {
		if _, e := VerifyV2(s[:i]); e == nil {
			t.Fatal(i)
		}
	}
}
func TestPersistentKey(t *testing.T) {
	path := t.TempDir() + "/key.pem"
	a, e := CreateIdentity(path)
	if e != nil {
		t.Fatal(e)
	}
	b, e := LoadIdentity(path)
	if e != nil || !bytes.Equal(a.Certificate, b.Certificate) || a.Key.N.Cmp(b.Key.N) != 0 {
		t.Fatal(e)
	}
	if _, e = CreateIdentity(path); e == nil {
		t.Fatal("existing identity overwritten")
	}
}
func FuzzVerify(f *testing.F) {
	f.Add([]byte("APK Sig Block 42"))
	f.Fuzz(func(t *testing.T, b []byte) { VerifyV2(b) })
}
func TestMultiChunkDigest(t *testing.T) {
	var buffer bytes.Buffer
	z := zip.NewWriter(&buffer)
	h := &zip.FileHeader{Name: "large", Method: zip.Store}
	w, e := z.CreateHeader(h)
	if e != nil {
		t.Fatal(e)
	}
	w.Write(bytes.Repeat([]byte("x"), 2<<20))
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	s, e := SignV2(buffer.Bytes(), testKey(t))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = VerifyV2(s); e != nil {
		t.Fatal(e)
	}
	s[1<<20] ^= 1
	if _, e = VerifyV2(s); e == nil {
		t.Fatal("multi-chunk content tampering")
	}
}
