package server

import "time"

const (
	defaultCacheDir        = "letsencrypt"
	defaultRefreshInterval = time.Hour * 24

	prefixBearer = "Bearer "

	certResourceFileName = "CertResource.json"

	certPemFileName = "cert.pem"

	keyPemFileName = "key.pem"
)
