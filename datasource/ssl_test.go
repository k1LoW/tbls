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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/k1LoW/tbls/config"
	"github.com/xo/dburl"
)

func TestApplyTLSConfigMySQL(t *testing.T) {
	dir := t.TempDir()
	ca := writeTestCA(t, dir)
	cert, key := writeTestKeyPair(t, dir)

	tests := []struct {
		name    string
		dsn     string
		tls     config.TLS
		wantErr bool
	}{
		{"ca only", "mysql://user:pass@hostname:3306/dbname", config.TLS{CA: ca}, false},
		{"ca with client cert", "mysql://user:pass@hostname:3306/dbname", config.TLS{CA: ca, Cert: cert, Key: key}, false},
		{"client cert only", "maria://user:pass@hostname:3306/dbname", config.TLS{Cert: cert, Key: key}, false},
		{"ca with verify identity", "mysql://user:pass@hostname:3306/dbname", config.TLS{CA: ca, Verify: "identity"}, false},
		{"verify identity only", "mysql://user:pass@hostname:3306/dbname", config.TLS{Verify: "identity"}, false},
		{"explicit tls=true keeps full verification", "mysql://user:pass@hostname:3306/dbname?tls=true", config.TLS{Cert: cert, Key: key}, false},
		{"explicit tls=skip-verify is upgraded", "mysql://user:pass@hostname:3306/dbname?tls=skip-verify", config.TLS{CA: ca}, false},
		{"explicit tls=preferred conflicts", "mysql://user:pass@hostname:3306/dbname?tls=preferred", config.TLS{CA: ca}, true},
		{"explicit tls=false conflicts", "mysql://user:pass@hostname:3306/dbname?tls=false", config.TLS{CA: ca}, true},
		{"invalid verify value", "mysql://user:pass@hostname:3306/dbname", config.TLS{CA: ca, Verify: "yeah"}, true},
		{"cert without key", "mysql://user:pass@hostname:3306/dbname", config.TLS{CA: ca, Cert: cert}, true},
		{"key without cert", "mysql://user:pass@hostname:3306/dbname", config.TLS{Key: key}, true},
		{"missing ca file", "mysql://user:pass@hostname:3306/dbname", config.TLS{CA: filepath.Join(dir, "missing.pem")}, true},
		{"invalid ca file", "mysql://user:pass@hostname:3306/dbname", config.TLS{CA: key}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := dburl.Parse(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			err = applyTLSConfig(u, tt.tls)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := u.Query().Get("tls"); !strings.HasPrefix(got, "tbls-dsn-") {
				t.Errorf("tls parameter should be a registered tbls-dsn config, got %q", got)
			}
		})
	}
}

func TestApplyTLSConfigMySQLRegistersUniqueNames(t *testing.T) {
	dir := t.TempDir()
	ca := writeTestCA(t, dir)
	names := map[string]bool{}
	for range 2 {
		u, err := dburl.Parse("mysql://user:pass@hostname:3306/dbname")
		if err != nil {
			t.Fatal(err)
		}
		if err := applyTLSConfig(u, config.TLS{CA: ca}); err != nil {
			t.Fatal(err)
		}
		names[u.Query().Get("tls")] = true
	}
	if len(names) != 2 {
		t.Errorf("expected unique tls config names per call, got %v", names)
	}
}

func TestApplyTLSConfigPostgres(t *testing.T) {
	dir := t.TempDir()
	ca := writeTestCA(t, dir)
	cert, key := writeTestKeyPair(t, dir)

	tests := []struct {
		name         string
		dsn          string
		tls          config.TLS
		wantErr      bool
		wantSSLMode  string
		wantRootCert string
		wantCert     string
		wantKey      string
	}{
		{"ca only", "postgres://user:pass@hostname:5432/dbname", config.TLS{CA: ca}, false, "verify-ca", ca, "", ""},
		{"ca keeps explicit verify-full", "postgres://user:pass@hostname:5432/dbname?sslmode=verify-full", config.TLS{CA: ca}, false, "verify-full", ca, "", ""},
		{"ca keeps explicit require", "postgres://user:pass@hostname:5432/dbname?sslmode=require", config.TLS{CA: ca}, false, "require", ca, "", ""},
		{"ca with client cert", "postgres://user:pass@hostname:5432/dbname", config.TLS{CA: ca, Cert: cert, Key: key}, false, "verify-ca", ca, cert, key},
		{"ca with verify identity", "postgres://user:pass@hostname:5432/dbname", config.TLS{CA: ca, Verify: "identity"}, false, "verify-full", ca, "", ""},
		{"redshift scheme", "rs://user:pass@hostname:5439/dbname", config.TLS{CA: ca}, false, "verify-ca", ca, "", ""},
		{"ca conflicts with sslmode=disable", "postgres://user:pass@hostname:5432/dbname?sslmode=disable", config.TLS{CA: ca}, true, "", "", "", ""},
		{"verify identity conflicts with explicit sslmode", "postgres://user:pass@hostname:5432/dbname?sslmode=require", config.TLS{Verify: "identity"}, true, "", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := dburl.Parse(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			err = applyTLSConfig(u, tt.tls)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
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

func TestApplyTLSConfigPostgresRejectsPathsWithSpaces(t *testing.T) {
	dir := t.TempDir()
	spacedDir := filepath.Join(dir, "with space")
	if err := os.Mkdir(spacedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	ca := writeTestCA(t, spacedDir)
	u, err := dburl.Parse("postgres://user:pass@hostname:5432/dbname")
	if err != nil {
		t.Fatal(err)
	}
	if err := applyTLSConfig(u, config.TLS{CA: ca}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApplyTLSConfigSQLServer(t *testing.T) {
	dir := t.TempDir()
	ca := writeTestCA(t, dir)
	cert, key := writeTestKeyPair(t, dir)

	t.Run("ca only", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance")
		if err != nil {
			t.Fatal(err)
		}
		if err := applyTLSConfig(u, config.TLS{CA: ca}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
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
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance")
		if err != nil {
			t.Fatal(err)
		}
		if err := applyTLSConfig(u, config.TLS{Verify: "identity"}); err != nil {
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
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?encrypt=strict")
		if err != nil {
			t.Fatal(err)
		}
		if err := applyTLSConfig(u, config.TLS{CA: ca}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := u.Query().Get("encrypt"); got != "strict" {
			t.Errorf("encrypt = %q, want %q", got, "strict")
		}
	})
	t.Run("ca conflicts with encrypt=disable", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?encrypt=disable")
		if err != nil {
			t.Fatal(err)
		}
		if err := applyTLSConfig(u, config.TLS{CA: ca}); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("ca conflicts with trustservercertificate=true", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?trustservercertificate=true")
		if err != nil {
			t.Fatal(err)
		}
		if err := applyTLSConfig(u, config.TLS{CA: ca}); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("verify identity conflicts with trustservercertificate=true", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance?trustservercertificate=true")
		if err != nil {
			t.Fatal(err)
		}
		if err := applyTLSConfig(u, config.TLS{Verify: "identity"}); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("client cert unsupported", func(t *testing.T) {
		u, err := dburl.Parse("sqlserver://user:pass@hostname:1433/instance")
		if err != nil {
			t.Fatal(err)
		}
		if err := applyTLSConfig(u, config.TLS{Cert: cert, Key: key}); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestApplyTLSConfigUnsupportedDriver(t *testing.T) {
	u, err := dburl.Parse("clickhouse://user:pass@hostname:9000/dbname")
	if err != nil {
		t.Fatal(err)
	}
	if err := applyTLSConfig(u, config.TLS{CA: "/path/to/ca.pem"}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAnalyzeTLSUnsupportedDatasource(t *testing.T) {
	for _, urlstr := range []string{
		"mongodb://user:pass@hostname:27017/dbname",
		"json://path/to/schema.json",
	} {
		if _, err := Analyze(config.DSN{URL: urlstr, TLS: config.TLS{CA: "/path/to/ca.pem"}}); err == nil {
			t.Errorf("expected error for %s, got nil", urlstr)
		}
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
