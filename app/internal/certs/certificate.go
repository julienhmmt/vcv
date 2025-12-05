package certs

import "time"

type Certificate struct {
	ID         string    `json:"id"`
	CommonName string    `json:"commonName"`
	Sans       []string  `json:"sans"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Revoked    bool      `json:"revoked"`
}

type DetailedCertificate struct {
	Certificate
	SerialNumber      string   `json:"serialNumber"`
	Issuer            string   `json:"issuer"`
	Subject           string   `json:"subject"`
	KeyAlgorithm      string   `json:"keyAlgorithm"`
	KeySize           int      `json:"keySize"`
	FingerprintSHA1   string   `json:"fingerprintSHA1"`
	FingerprintSHA256 string   `json:"fingerprintSHA256"`
	Usage             []string `json:"usage"`
	PEM               string   `json:"pem"`
}

type PEMResponse struct {
	SerialNumber string `json:"serialNumber"`
	PEM          string `json:"pem"`
}
