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
