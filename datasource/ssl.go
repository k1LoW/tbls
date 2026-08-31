package datasource

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/k1LoW/tbls/config"
	"github.com/xo/dburl"
)

var mysqlTLSConfigCounter atomic.Uint64

// tlsSupportedDrivers are the dburl drivers that the dsn.tls configuration can
// be applied to.
var tlsSupportedDrivers = []string{"mysql", "postgres", "sqlserver"}

const tlsVerifyIdentity = "identity"

// validateTLSOptions checks the dsn.tls fields that every driver shares, so
// that the driver-specific appliers only deal with their own parameters.
func validateTLSOptions(t config.TLS) error {
	if t.Verify != "" && t.Verify != tlsVerifyIdentity {
		return fmt.Errorf("invalid dsn.tls.verify value: %s", t.Verify)
	}
	if (t.Cert == "") != (t.Key == "") {
		return fmt.Errorf("dsn.tls.cert and dsn.tls.key must be set together")
	}
	return nil
}

// applyTLSConfig applies the dsn.tls configuration to the connection using
// each driver's own TLS mechanism, rewriting the DSN query accordingly.
// Empty values are treated as absent. Native TLS DSN parameters that
// contradict the configuration (e.g. sslmode=disable, encrypt=disable,
// trustservercertificate=true) are rejected instead of silently overridden.
func applyTLSConfig(u *dburl.URL, t config.TLS) error {
	if err := validateTLSOptions(t); err != nil {
		return err
	}

	values := u.Query()
	var err error
	switch u.Driver {
	case "mysql":
		err = registerMySQLTLS(t, values)
	case "postgres":
		err = applyPostgresTLS(t, values)
	case "sqlserver":
		err = applySQLServerTLS(t, values)
	default:
		err = fmt.Errorf("dsn.tls is not supported for driver '%s'", u.Driver)
	}
	if err != nil {
		return err
	}
	u.RawQuery = values.Encode()
	return nil
}

// registerMySQLTLS registers a TLS config for MySQL/MariaDB connections and
// rewrites the tls parameter to use it. An explicit tls=true is honored by
// keeping full verification; tls=false and tls=preferred cannot be combined
// with dsn.tls and are rejected.
func registerMySQLTLS(t config.TLS, values url.Values) error {
	identity := t.Verify == tlsVerifyIdentity
	switch explicit := values.Get("tls"); explicit {
	case "", "skip-verify":
	case "true":
		identity = true
	default:
		return fmt.Errorf("cannot combine tls=%s with dsn.tls", explicit)
	}

	tlsConfig := &tls.Config{} //nolint:gosec // the verification level is set explicitly below
	var pool *x509.CertPool
	if t.CA != "" {
		pem, err := os.ReadFile(t.CA)
		if err != nil {
			return fmt.Errorf("failed to read dsn.tls.ca: %w", err)
		}
		pool = x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return fmt.Errorf("failed to parse CA certificate from dsn.tls.ca: %s", t.CA)
		}
	}
	switch {
	case identity:
		// Full verification (chain + hostname) against dsn.tls.ca, or the system roots when no ca is given.
		tlsConfig.RootCAs = pool
	case pool != nil:
		// VERIFY_CA: chain verification without hostname verification.
		tlsConfig.InsecureSkipVerify = true
		tlsConfig.VerifyConnection = verifyServerCertificate(pool)
	default:
		// Encrypted without server verification, like tls=skip-verify.
		tlsConfig.InsecureSkipVerify = true
	}
	if t.Cert != "" {
		pair, err := tls.LoadX509KeyPair(t.Cert, t.Key)
		if err != nil {
			return fmt.Errorf("failed to load dsn.tls.cert/dsn.tls.key: %w", err)
		}
		// GetClientCertificate (instead of Certificates) sends the certificate
		// even when the server's acceptable-CA list does not match its issuer.
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &pair, nil
		}
	}

	name := fmt.Sprintf("tbls-dsn-%d", mysqlTLSConfigCounter.Add(1))
	if err := mysqldriver.RegisterTLSConfig(name, tlsConfig); err != nil {
		return err
	}
	values.Set("tls", name)
	return nil
}

// applyPostgresTLS maps the dsn.tls configuration to lib/pq's native ssl
// parameters.
func applyPostgresTLS(t config.TLS, values url.Values) error {
	for _, path := range []string{t.CA, t.Cert, t.Key} {
		if strings.Contains(path, " ") {
			return fmt.Errorf("certificate paths containing spaces are not supported for postgres DSNs: %s", path)
		}
	}
	mode := values.Get("sslmode")
	switch mode {
	case "disable", "allow", "prefer":
		return fmt.Errorf("dsn.tls conflicts with sslmode=%s", mode)
	}
	if t.Verify == tlsVerifyIdentity {
		if mode != "" && mode != "verify-full" {
			return fmt.Errorf("dsn.tls.verify: identity conflicts with sslmode=%s", mode)
		}
		values.Set("sslmode", "verify-full")
	} else if t.CA != "" && mode == "" {
		values.Set("sslmode", "verify-ca")
	}
	if t.CA != "" {
		values.Set("sslrootcert", t.CA)
	}
	if t.Cert != "" {
		values.Set("sslcert", t.Cert)
		values.Set("sslkey", t.Key)
	}
	return nil
}

// applySQLServerTLS maps dsn.tls.ca to go-mssqldb's certificate parameter (the
// file must have a .pem or .der extension). go-mssqldb performs full
// verification (chain + hostname) when encryption is on and
// trustservercertificate is false, so verify: identity needs no extra mapping.
func applySQLServerTLS(t config.TLS, values url.Values) error {
	if t.Cert != "" {
		return fmt.Errorf("dsn.tls.cert/dsn.tls.key are not supported for driver 'sqlserver'")
	}
	if v := values.Get("trustservercertificate"); v != "" {
		trust, err := strconv.ParseBool(v)
		if err == nil && trust {
			return fmt.Errorf("dsn.tls conflicts with trustservercertificate=true")
		}
	}
	switch strings.ToLower(values.Get("encrypt")) {
	case "":
		values.Set("encrypt", "true")
	case "disable", "optional", "no", "0", "f", "false":
		return fmt.Errorf("dsn.tls conflicts with encrypt=%s", values.Get("encrypt"))
	}
	if t.CA != "" {
		values.Set("certificate", t.CA)
		values.Set("trustservercertificate", "false")
	}
	return nil
}

// verifyServerCertificate verifies the server chain against roots, skipping
// hostname verification. VerifyConnection (instead of VerifyPeerCertificate)
// also covers resumed sessions.
func verifyServerCertificate(roots *x509.CertPool) func(tls.ConnectionState) error {
	return func(cs tls.ConnectionState) error {
		if len(cs.PeerCertificates) == 0 {
			return fmt.Errorf("no server certificate presented")
		}
		opts := x509.VerifyOptions{
			Roots:         roots,
			Intermediates: x509.NewCertPool(),
		}
		for _, cert := range cs.PeerCertificates[1:] {
			opts.Intermediates.AddCert(cert)
		}
		if _, err := cs.PeerCertificates[0].Verify(opts); err != nil {
			return fmt.Errorf("failed to verify server certificate against dsn.tls.ca: %w", err)
		}
		return nil
	}
}
