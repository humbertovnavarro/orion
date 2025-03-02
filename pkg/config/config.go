package config

import (
	"os"

	"github.com/joho/godotenv"
)

var CertificateAuthorityLocation string
var CertificateAuthorityKeyLocation string
var CertificateDuration string
var NebulaBin string
var NebulaCertBin string
var ConfigLocation string
var ClientConfigLocation string

func init() {
	LoadEnvironment()
}

func LoadEnvironment() {
	godotenv.Load()
	ClientConfigLocation = load("NEBULA_CONFIG_LOCATION", "/etc/nebula/clients")
	ConfigLocation = load("NEBULA_CONFIG_LOCATION", "/etc/nebula")
	CertificateAuthorityLocation = load("CERTIFICATE_LOCATION", "/etc/nebula/orion.crt")
	CertificateAuthorityKeyLocation = load("CERTIFICATE_KEY_LOCATION", "/etc/nebula/orion.key")
	CertificateDuration = load("CERTIFICATE_DURATION", "1h")
	NebulaBin = load("NEBULA_BIN", "/usr/bin/nebula")
	NebulaCertBin = load("NEBULA_CERT_BIN", "/usr/bin/nebula-cert")
	os.MkdirAll(ConfigLocation, 0644)
	os.MkdirAll(ClientConfigLocation, 0644)
}

func load(env string, def string) (out string) {
	out = os.Getenv(env)
	if out == "" {
		return def
	}
	return
}
