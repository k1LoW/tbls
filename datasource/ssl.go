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
	"github.com/xo/dburl"
)

var mysqlTLSConfigCounter atomic.Uint64

var sslParamKeys = []string{"ssl-ca", "ssl-cert", "ssl-key", "ssl-verify-identity"}

type sslParams struct {
	ca             string
	cert           string
	key            string
	verifyIdentity bool
}

// applySSLParams extracts the ssl-ca, ssl-cert, ssl-key and ssl-verify-identity
// DSN parameters and applies them to the connection using each driver's own TLS
// mechanism. Empty values are treated as absent. Explicit native TLS parameters
// that contradict them (e.g. sslmode=disable, encrypt=disable,
// trustservercertificate=true) are rejected instead of silently overridden.
// Reports whether the DSN query was rewritten.
func applySSLParams(u *dburl.URL) (bool, error) {
	values := u.Query()
	present := false
	for _, k := range sslParamKeys {
		if _, ok := values[k]; ok {
			present = true
		}
	}
	if !present {
		return false, nil
	}
	p := sslParams{
		ca:   values.Get("ssl-ca"),
		cert: values.Get("ssl-cert"),
		key:  values.Get("ssl-key"),
	}
	if v := values.Get("ssl-verify-identity"); v != "" {
		verifyIdentity, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("invalid ssl-verify-identity value: %s", v)
		}
		p.verifyIdentity = verifyIdentity
	}
	for _, k := range sslParamKeys {
		values.Del(k)
	}

	if p.ca == "" && p.cert == "" && p.key == "" && !p.verifyIdentity {
		u.RawQuery = values.Encode()
		return true, nil
	}
	if (p.cert == "") != (p.key == "") {
		return false, fmt.Errorf("ssl-cert and ssl-key must be set together")
	}

	var err error
	switch u.Driver {
	case "mysql":
		err = registerMySQLTLS(p, values)
	case "postgres":
		err = applyPostgresSSL(p, values)
	case "sqlserver":
		err = applySQLServerSSL(p, values)
	default:
		err = fmt.Errorf("ssl-ca/ssl-cert/ssl-key/ssl-verify-identity are not supported for driver '%s'", u.Driver)
	}
	if err != nil {
		return false, err
	}
	u.RawQuery = values.Encode()
	return true, nil
}

// registerMySQLTLS registers a TLS config for MySQL/MariaDB connections and
// rewrites the tls parameter to use it. An explicit tls=true is honored by
// keeping full verification; tls=false and tls=preferred cannot be combined
// with certificate files and are rejected.
func registerMySQLTLS(p sslParams, values url.Values) error {
	switch explicit := values.Get("tls"); explicit {
	case "", "skip-verify":
	case "true":
		p.verifyIdentity = true
	default:
		return fmt.Errorf("cannot combine tls=%s with ssl-ca/ssl-cert/ssl-key", explicit)
	}

	tlsConfig := &tls.Config{} //nolint:gosec // the verification level is set explicitly below
	var pool *x509.CertPool
	if p.ca != "" {
		pem, err := os.ReadFile(p.ca)
		if err != nil {
			return fmt.Errorf("failed to read ssl-ca: %w", err)
		}
		pool = x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return fmt.Errorf("failed to parse CA certificate from ssl-ca: %s", p.ca)
		}
	}
	switch {
	case p.verifyIdentity:
		// Full verification (chain + hostname) against ssl-ca, or the system roots when no ssl-ca is given.
		tlsConfig.RootCAs = pool
	case pool != nil:
		// VERIFY_CA: chain verification without hostname verification.
		tlsConfig.InsecureSkipVerify = true
		tlsConfig.VerifyConnection = verifyServerCertificate(pool)
	default:
		// Encrypted without server verification, like tls=skip-verify.
		tlsConfig.InsecureSkipVerify = true
	}
	if p.cert != "" {
		pair, err := tls.LoadX509KeyPair(p.cert, p.key)
		if err != nil {
			return fmt.Errorf("failed to load ssl-cert/ssl-key: %w", err)
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

// applyPostgresSSL maps the parameters to lib/pq's native ssl parameters.
func applyPostgresSSL(p sslParams, values url.Values) error {
	for _, path := range []string{p.ca, p.cert, p.key} {
		if strings.Contains(path, " ") {
			return fmt.Errorf("certificate paths containing spaces are not supported for postgres DSNs: %s", path)
		}
	}
	mode := values.Get("sslmode")
	switch mode {
	case "disable", "allow", "prefer":
		return fmt.Errorf("ssl-ca/ssl-cert/ssl-key/ssl-verify-identity conflict with sslmode=%s", mode)
	}
	if p.verifyIdentity {
		if mode != "" && mode != "verify-full" {
			return fmt.Errorf("ssl-verify-identity=true conflicts with sslmode=%s", mode)
		}
		values.Set("sslmode", "verify-full")
	} else if p.ca != "" && mode == "" {
		values.Set("sslmode", "verify-ca")
	}
	if p.ca != "" {
		values.Set("sslrootcert", p.ca)
	}
	if p.cert != "" {
		values.Set("sslcert", p.cert)
		values.Set("sslkey", p.key)
	}
	return nil
}

// applySQLServerSSL maps ssl-ca to go-mssqldb's certificate parameter (the file
// must have a .pem or .der extension). go-mssqldb performs full verification
// (chain + hostname) when encryption is on and trustservercertificate is
// false, so ssl-verify-identity needs no extra mapping.
func applySQLServerSSL(p sslParams, values url.Values) error {
	if p.cert != "" {
		return fmt.Errorf("ssl-cert/ssl-key are not supported for driver 'sqlserver'")
	}
	if t := values.Get("trustservercertificate"); t != "" {
		trust, err := strconv.ParseBool(t)
		if err == nil && trust && p.ca != "" {
			return fmt.Errorf("ssl-ca conflicts with trustservercertificate=true")
		}
	}
	switch strings.ToLower(values.Get("encrypt")) {
	case "":
		values.Set("encrypt", "true")
	case "disable", "optional", "no", "0", "f", "false":
		return fmt.Errorf("ssl-ca/ssl-verify-identity conflict with encrypt=%s", values.Get("encrypt"))
	}
	if p.ca != "" {
		values.Set("certificate", p.ca)
		values.Set("trustservercertificate", "false")
	}
	return nil
}

// rejectSSLParams returns an error when the DSN carries generic ssl parameters
// for a datasource that does not support them, instead of silently ignoring
// the requested security settings.
func rejectSSLParams(urlstr string) error {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil //nolint:nilerr // let the datasource report its own DSN parse error
	}
	for _, k := range sslParamKeys {
		if _, ok := u.Query()[k]; ok {
			return fmt.Errorf("ssl-ca/ssl-cert/ssl-key/ssl-verify-identity are not supported for datasource '%s://'", u.Scheme)
		}
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
			return fmt.Errorf("failed to verify server certificate against ssl-ca: %w", err)
		}
		return nil
	}
}
