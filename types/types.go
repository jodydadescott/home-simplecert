package types

import (
	"net/http"
	"net/http/httputil"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/jinzhu/copier"
	hashserver "github.com/jodydadescott/simple-go-hash-auth/server"
)

type CR struct {
	Domain            string                `json:"domain,omitempty" yaml:"domain,omitempty"`
	CertURL           string                `json:"certUrl,omitempty" yaml:"certUrl,omitempty"`
	CertStableURL     string                `json:"certStableUrl,omitempty" yaml:"certStableUrl,omitempty"`
	PrivateKey        []byte                `json:"privateKey,omitempty" yaml:"privateKey,omitempty"`
	Certificate       []byte                `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	IssuerCertificate []byte                `json:"issuerCertificate,omitempty" yaml:"issuerCertificate,omitempty"`
	CSR               []byte                `json:"csr,omitempty" yaml:"csr,omitempty"`
	certResource      *certificate.Resource `json:"-"`
}

// Clone return copy
func (t *CR) Clone() *CR {
	c := &CR{}
	copier.Copy(&c, &t)
	return c
}

func (t *CR) initCertResource() {
	if t.certResource != nil {
		return
	}
	t.certResource = &certificate.Resource{}
	t.certResource.Domain = t.Domain
	t.certResource.CertURL = t.CertURL
	t.certResource.CertStableURL = t.CertStableURL
	t.certResource.PrivateKey = t.PrivateKey
	t.certResource.Certificate = t.Certificate
	t.certResource.IssuerCertificate = t.IssuerCertificate
	t.certResource.CSR = t.CSR
}

func (t *CR) GetCertPEM() []byte {
	t.initCertResource()
	return t.certResource.Certificate
}

func (t *CR) GetKeyPEM() []byte {
	t.initCertResource()
	return t.certResource.PrivateKey
}

type TokenResponse struct {
	*hashserver.Token
	Error string
}

// Clone return copy
func (t *TokenResponse) Clone() *TokenResponse {
	c := &TokenResponse{}
	copier.Copy(&c, &t)
	return c
}

type CertResponse struct {
	CR    *CR    `json:"cr,omitempty" yaml:"cr,omitempty"`
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Clone return copy
func (t *CertResponse) Clone() *CertResponse {
	c := &CertResponse{}
	copier.Copy(&c, &t)
	return c
}

type SimpleMessage struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Clone return copy
func (t *SimpleMessage) Clone() *SimpleMessage {
	c := &SimpleMessage{}
	copier.Copy(&c, &t)
	return c
}

type HTTPDebug struct {
	Request  string `json:"request,omitempty" yaml:"request,omitempty"`
	Response string `json:"response,omitempty" yaml:"response,omitempty"`
}

func NewHTTPDebug(r *http.Request, response []byte) *HTTPDebug {

	breq, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}

	return &HTTPDebug{
		Request:  string(breq),
		Response: string(response),
	}
}

// Clone return copy
func (t *HTTPDebug) Clone() *HTTPDebug {
	c := &HTTPDebug{}
	copier.Copy(&c, &t)
	return c
}
