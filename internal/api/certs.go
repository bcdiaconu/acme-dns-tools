package api

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// allowedCertFiles lists the only file names that may be served.
// This prevents enumeration / traversal of arbitrary files.
var allowedCertFiles = map[string]bool{
	"fullchain.pem": true,
	"privkey.pem":   true,
	"cert.pem":      true,
	"chain.pem":     true,
}

// CertsHandler returns an http.HandlerFunc that serves certificate files from
// certsBaseDir (typically /etc/letsencrypt/live) under the path
//
//	GET /certs/{domain}/{file}
//
// Authentication:
//   - Bearer token check (Authorization: Bearer <token>)
//   - Forward-Confirmed Reverse DNS (FCrDNS) allowlist:
//     client IP → PTR → A/AAAA → confirm original IP is present AND
//     the resolved hostname is in dnsAllowlist.
func CertsHandler(bearerToken string, dnsAllowlist []string, certsBaseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// --- Bearer token ---
		if r.Header.Get("Authorization") != "Bearer "+bearerToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// --- FCrDNS allowlist ---
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("certs: cannot parse RemoteAddr %q: %v", r.RemoteAddr, err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if !isAllowedByFCrDNS(clientIP, dnsAllowlist) {
			log.Printf("certs: denied request from %s – not in DNS allowlist", clientIP)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// --- Parse /certs/{domain}/{file} ---
		// http.ServeMux strips the registered prefix but we registered "/certs/",
		// so r.URL.Path still contains the full path.
		trimmed := strings.TrimPrefix(r.URL.Path, "/certs/")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			http.Error(w, "Bad Request – expected /certs/{domain}/{file}", http.StatusBadRequest)
			return
		}
		domain := parts[0]
		fileName := parts[1]

		// --- Validate domain (no path traversal) ---
		if strings.Contains(domain, "..") || strings.Contains(domain, "/") || strings.Contains(domain, "\\") {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// --- Validate file name (allowlist only) ---
		if !allowedCertFiles[fileName] {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// --- Read file ---
		// filepath.Join is safe here because domain and fileName are already validated.
		certPath := filepath.Join(certsBaseDir, domain, fileName)
		data, err := os.ReadFile(certPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				log.Printf("certs: failed to read %s: %v", certPath, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		log.Printf("certs: served %s to %s", certPath, clientIP)
		w.Header().Set("Content-Type", "application/x-pem-file")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// isAllowedByFCrDNS performs Forward-Confirmed Reverse DNS verification:
//  1. Reverse lookup: clientIP → PTR record(s) → hostname(s)
//  2. For each hostname in the DNS allowlist: forward lookup → IPs
//  3. Allow only if the original clientIP appears in the forward IPs.
//
// This prevents spoofing via arbitrary PTR records: the admin must also control
// the forward (A/AAAA) DNS for the allowed hostname.
func isAllowedByFCrDNS(clientIP string, allowlist []string) bool {
	if len(allowlist) == 0 {
		return false
	}

	// Reverse lookup
	ptrs, err := net.LookupAddr(clientIP)
	if err != nil || len(ptrs) == 0 {
		return false
	}

	for _, ptr := range ptrs {
		// net.LookupAddr returns FQDNs with a trailing dot
		hostname := strings.TrimSuffix(ptr, ".")

		for _, allowed := range allowlist {
			if !strings.EqualFold(hostname, allowed) {
				continue
			}
			// Forward-confirm: resolve the allowed hostname → check clientIP is present
			addrs, err := net.LookupHost(hostname)
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if addr == clientIP {
					return true
				}
			}
		}
	}
	return false
}
