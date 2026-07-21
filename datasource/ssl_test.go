package datasource

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xo/dburl"
)

func TestApplySSLParamsAbsent(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		wantChanged bool
	}{
		{"no ssl params", "mysql://user:pass@hostname:3306/dbname?hide_auto_increment", false},
		{"empty ssl params", "mysql://user:pass@hostname:3306/dbname?ssl-ca=&ssl-cert=&ssl-key=&ssl-verify-identity=", true},
		{"empty ssl params on postgres", "postgres://user:pass@hostname:5432/dbname?ssl-ca=", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := dburl.Parse(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := applySSLParams(u)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if changed != tt.wantChanged {
				t.Errorf("changed = %v, want %v", changed, tt.wantChanged)
			}
			assertSSLParamsRemoved(t, u)
			if got := u.Query().Get("tls"); got != "" {
				t.Errorf("tls parameter should not be set, got %q", got)
			}
		})
	}
}

func TestApplySSLParamsMySQL(t *testing.T) {
	dir := t.TempDir()
	ca := writeTestCA(t, dir)
	cert, key := writeTestKeyPair(t, dir)

	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{"ca only", "mysql://user:pass@hostname:3306/dbname?ssl-ca=" + ca, false},
		{"ca with client cert", "mysql://user:pass@hostname:3306/dbname?ssl-ca=" + ca + "&ssl-cert=" + cert + "&ssl-key=" + key, false},
		{"client cert only", "maria://user:pass@hostname:3306/dbname?ssl-cert=" + cert + "&ssl-key=" + key, false},
		{"ca with verify identity", "mysql://user:pass@hostname:3306/dbname?ssl-ca=" + ca + "&ssl-verify-identity=true", false},
		{"verify identity only", "mysql://user:pass@hostname:3306/dbname?ssl-verify-identity=true", false},
		{"explicit tls=true keeps full verification", "mysql://user:pass@hostname:3306/dbname?tls=true&ssl-cert=" + cert + "&ssl-key=" + key, false},
		{"explicit tls=skip-verify is upgraded", "mysql://user:pass@hostname:3306/dbname?tls=skip-verify&ssl-ca=" + ca, false},
		{"explicit tls=preferred conflicts", "mysql://user:pass@hostname:3306/dbname?tls=preferred&ssl-ca=" + ca, true},
		{"explicit tls=false conflicts", "mysql://user:pass@hostname:3306/dbname?tls=false&ssl-ca=" + ca, true},
		{"invalid verify identity value", "mysql://user:pass@hostname:3306/dbname?ssl-ca=" + ca + "&ssl-verify-identity=yeah", true},
		{"cert without key", "mysql://user:pass@hostname:3306/dbname?ssl-ca=" + ca + "&ssl-cert=" + cert, true},
		{"key without cert", "mysql://user:pass@hostname:3306/dbname?ssl-key=" + key, true},
		{"missing ca file", "mysql://user:pass@hostname:3306/dbname?ssl-ca=" + filepath.Join(dir, "missing.pem"), true},
		{"invalid ca file", "mysql://user:pass@hostname:3306/dbname?ssl-ca=" + key, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := dburl.Parse(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := applySSLParams(u)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !changed {
				t.Error("changed = false, want true")
			}
			assertSSLParamsRemoved(t, u)
			if got := u.Query().Get("tls"); !strings.HasPrefix(got, "tbls-dsn-") {
				t.Errorf("tls parameter should be a registered tbls-dsn config, got %q", got)
			}
		})
	}
}

func TestApplySSLParamsMySQLRegistersUniqueNames(t *testing.T) {
	dir := t.TempDir()
	ca := writeTestCA(t, dir)
	names := map[string]bool{}
	for range 2 {
		u, err := dburl.Parse("mysql://user:pass@hostname:3306/dbname?ssl-ca=" + ca)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := applySSLParams(u); err != nil {
			t.Fatal(err)
		}
		names[u.Query().Get("tls")] = true
	}
	if len(names) != 2 {
		t.Errorf("expected unique tls config names per call, got %v", names)
	}
}

func TestApplySSLParamsPostgres(t *testing.T) {
	dir := t.TempDir()
	ca := writeTestCA(t, dir)
	cert, key := writeTestKeyPair(t, dir)

	tests := []struct {
		name         string
		dsn          string
		wantErr      bool
		wantSSLMode  string
		wantRootCert string
		wantCert     string
		wantKey      string
	}{
		{"ca only", "postgres://user:pass@hostname:5432/dbname?ssl-ca=" + ca, false, "verify-ca", ca, "", ""},
		{"ca keeps explicit verify-full", "postgres://user:pass@hostname:5432/dbname?sslmode=verify-full&ssl-ca=" + ca, false, "verify-full", ca, "", ""},
		{"ca keeps explicit require", "postgres://user:pass@hostname:5432/dbname?sslmode=require&ssl-ca=" + ca, false, "require", ca, "", ""},
		{"ca with client cert", "postgres://user:pass@hostname:5432/dbname?ssl-ca=" + ca + "&ssl-cert=" + cert + "&ssl-key=" + key, false, "verify-ca", ca, cert, key},
		{"ca with verify identity", "postgres://user:pass@hostname:5432/dbname?ssl-ca=" + ca + "&ssl-verify-identity=true", false, "verify-full", ca, "", ""},
		{"redshift scheme", "rs://user:pass@hostname:5439/dbname?ssl-ca=" + ca, false, "verify-ca", ca, "", ""},
		{"ca conflicts with sslmode=disable", "postgres://user:pass@hostname:5432/dbname?sslmode=disable&ssl-ca=" + ca, true, "", "", "", ""},
		{"verify identity conflicts with explicit sslmode", "postgres://user:pass@hostname:5432/dbname?sslmode=require&ssl-verify-identity=true", true, "", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := dburl.Parse(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := applySSLParams(u)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !changed {
				t.Error("changed = false, want true")
			}
			assertSSLParamsRemoved(t, u)
			values := u.Query()
			for param, want := range map[string]string{
				"sslmode":     tt.wantSSLMode,
				"sslrootcert": tt.wantRootCert,
				"sslcert":     tt.wantCert,
				"sslkey":      tt.wantKey,
			} {
				if got := values.Get(param); got != want {
					t.Errorf("%s = %q, want %q", param, got, want)
				}
			}
		})
	}
}

func TestApplySSLParamsPostgresRejectsPathsWithSpaces(t *testing.T) {
	dir := t.TempDir()
	spacedDir := filepath.Join(dir, "with space")
	if err := os.Mkdir(spacedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	ca := writeTestCA(t, spacedDir)
	u, err := dburl.Parse("postgres://user:pass@hostname:5432/dbname?ssl-ca=" + url.QueryEscape(ca))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := applySSLParams(u); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApplySSLParamsSQLServer(t *testing.T) {
	dir := t.TempDir()
	ca := writeTestCA(t, dir)
	cert, key := writeTestKeyPair(t, dir)

	t.Run("ca only", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?ssl-ca=" + ca)
		if err != nil {
			t.Fatal(err)
		}
		changed, err := applySSLParams(u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !changed {
			t.Error("changed = false, want true")
		}
		assertSSLParamsRemoved(t, u)
		values := u.Query()
		if got := values.Get("encrypt"); got != "true" {
			t.Errorf("encrypt = %q, want %q", got, "true")
		}
		if got := values.Get("certificate"); got != ca {
			t.Errorf("certificate = %q, want %q", got, ca)
		}
		if got := values.Get("trustservercertificate"); got != "false" {
			t.Errorf("trustservercertificate = %q, want %q", got, "false")
		}
	})
	t.Run("verify identity only", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?ssl-verify-identity=true")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := applySSLParams(u); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		values := u.Query()
		if got := values.Get("encrypt"); got != "true" {
			t.Errorf("encrypt = %q, want %q", got, "true")
		}
		if _, ok := values["certificate"]; ok {
			t.Error("certificate should not be set")
		}
	})
	t.Run("explicit encrypt=strict is kept", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?encrypt=strict&ssl-ca=" + ca)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := applySSLParams(u); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := u.Query().Get("encrypt"); got != "strict" {
			t.Errorf("encrypt = %q, want %q", got, "strict")
		}
	})
	t.Run("ca conflicts with encrypt=disable", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?encrypt=disable&ssl-ca=" + ca)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := applySSLParams(u); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("ca conflicts with trustservercertificate=true", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?trustservercertificate=true&ssl-ca=" + ca)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := applySSLParams(u); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("client cert unsupported", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?ssl-cert=" + cert + "&ssl-key=" + key)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := applySSLParams(u); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestApplySSLParamsUnsupportedDriver(t *testing.T) {
	u, err := dburl.Parse("clickhouse://user:pass@hostname:9000/dbname?ssl-ca=/path/to/ca.pem")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := applySSLParams(u); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRejectSSLParams(t *testing.T) {
	if err := rejectSSLParams("mongodb://user:pass@hostname:27017/dbname?ssl-ca=/path/to/ca.pem"); err == nil {
		t.Error("expected error, got nil")
	}
	if err := rejectSSLParams("mongodb://user:pass@hostname:27017/dbname?tls=true"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyServerCertificate(t *testing.T) {
	caCert, caKey := generateCA(t)
	leaf := generateLeaf(t, caCert, caKey)
	otherCACert, otherCAKey := generateCA(t)
	otherLeaf := generateLeaf(t, otherCACert, otherCAKey)

	pool := x509.NewCertPool()
	pool.AddCert(caCert)
	verify := verifyServerCertificate(pool)

	if err := verify(tls.ConnectionState{PeerCertificates: []*x509.Certificate{leaf}}); err != nil {
		t.Errorf("valid chain rejected: %v", err)
	}
	if err := verify(tls.ConnectionState{PeerCertificates: []*x509.Certificate{otherLeaf}}); err == nil {
		t.Error("certificate from a different CA accepted")
	}
	if err := verify(tls.ConnectionState{}); err == nil {
		t.Error("connection without server certificate accepted")
	}
}

func assertSSLParamsRemoved(t *testing.T, u *dburl.URL) {
	t.Helper()
	values := u.Query()
	for _, k := range sslParamKeys {
		if _, ok := values[k]; ok {
			t.Errorf("%s should be removed from the DSN", k)
		}
	}
}

func generateCA(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "tbls test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert, priv
}

func generateLeaf(t *testing.T, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *x509.Certificate {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "tbls test server"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &priv.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}

func writeTestCA(t *testing.T, dir string) string {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "tbls test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "ca.pem")
	writePEM(t, path, "CERTIFICATE", der)
	return path
}

func writeTestKeyPair(t *testing.T, dir string) (string, string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "tbls test client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	certPath := filepath.Join(dir, "client-cert.pem")
	keyPath := filepath.Join(dir, "client-key.pem")
	writePEM(t, certPath, "CERTIFICATE", der)
	writePEM(t, keyPath, "EC PRIVATE KEY", keyDER)
	return certPath, keyPath
}

func writePEM(t *testing.T, path, blockType string, der []byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: blockType, Bytes: der}); err != nil {
		t.Fatal(err)
	}
}
