package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/foomo/simplecert"
	logger "github.com/jodydadescott/jody-go-logger"
	hashserver "github.com/jodydadescott/simple-go-hash-auth/server"
	"go.uber.org/zap"

	"github.com/jodydadescott/home-simplecert/types"
)

type DomainWrapper struct {
	sync.RWMutex
	*Domain
	cr  *CR
	err error
	*Server
}

func (t *DomainWrapper) init() error {

	if logger.Trace {
		zap.L().Debug("func (t *DomainWrapper) init() error")
	}

	cacheDir := filepath.Join(t.cacheDir, t.Name)
	if logger.Trace {
		zap.L().Debug(fmt.Sprintf("CacheDir is %s", cacheDir))
	}

	load := func() error {

		t.Lock()
		defer t.Unlock()

		b, err := os.ReadFile(filepath.Join(cacheDir, certResourceFileName))
		if err != nil {
			t.err = err
			zap.L().Error(fmt.Sprintf("Processing domain %s had error %s", t.Name, err.Error()))
			return err
		}

		cr := &CR{}
		err = json.Unmarshal(b, cr)
		if err != nil {
			t.err = err
			zap.L().Error(fmt.Sprintf("Processing domain %s had error %s", t.Name, err.Error()))
			return err
		}
		t.cr = cr

		return nil
	}

	setErr := func(err error) {
		t.Lock()
		defer t.Unlock()
		t.err = err
	}

	cfg := &simplecert.Config{
		CacheDir:      cacheDir,
		RenewBefore:   30 * 24,
		CheckInterval: 2 * 24 * time.Hour,
		DirectoryURL:  "https://acme-v02.api.letsencrypt.org/directory",
		HTTPAddress:   ":80",
		TLSAddress:    ":443",
		CacheDirPerm:  0700,
		SSLEmail:      t.email,
		KeyType:       simplecert.RSA2048,
		WillRenewCertificate: func() {
			zap.L().Info(fmt.Sprintf("Renewing domain %s", t.Name))
			t.stopServer()
		},
		DidRenewCertificate: func() {
			zap.L().Info(fmt.Sprintf("Renewed domain %s", t.Name))
			load()
			t.startServer()
		},
		FailedToRenewCertificate: func(err error) {

			setErr(err)

			if t.Name == t.primaryDomain {
				zap.L().Error(fmt.Sprintf("Failed to renew primary domain %s; error %s", t.Name, err.Error()))
			} else {
				zap.L().Error(fmt.Sprintf("Failed to renew domain %s; error %s", t.Name, err.Error()))
			}

			t.startServer()

		},
	}

	cfg.Domains = append(cfg.Domains, t.Name)

	if len(t.Aliases) > 0 {
		cfg.Domains = append(cfg.Domains, t.Aliases...)
	}

	if logger.Trace {
		zap.L().Debug(fmt.Sprintf("Initializing SimpleCert for domain %s", t.Name))
	}

	t.wg.Add(1)

	if logger.Trace {
		zap.L().Debug("t.wg.Add(1)")
	}

	_, err := simplecert.Init(cfg, func() {

		if logger.Trace {
			zap.L().Debug("t.wg.Done()")
		}

		t.wg.Done()

		zap.L().Debug(fmt.Sprintf("Closing SimpleCert for domain %s", t.Name))
	})

	if err != nil {
		t.err = err
		zap.L().Error(fmt.Sprintf("Processing domain %s had error %s", t.Name, err.Error()))
		return err
	}

	return load()
}

func (t *DomainWrapper) get() (*CR, error) {
	t.RLock()
	defer t.RUnlock()
	return t.cr, t.err
}

type Server struct {
	primaryDomain string
	domains       map[string]*DomainWrapper
	email         string
	cacheDir      string
	hashserver    *hashserver.Server
	mutex         sync.Mutex
	cancel        context.CancelFunc
	errc          chan error
	embargo       bool
	wg            sync.WaitGroup
}

func New(config *Config) (*Server, error) {

	if logger.Trace {
		zap.L().Debug("New(config *Config) (*Server, error)")
	}

	if config == nil {
		panic("config is nil")
	}

	config = config.Clone()

	if config.Secret == "" {
		return nil, fmt.Errorf("secret is required")
	}

	if config.PrimaryDomain == nil {
		return nil, fmt.Errorf("primary domain is required")
	}

	if config.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if config.CacheDir == "" {
		config.CacheDir = defaultCacheDir
	}

	s := &Server{
		domains: make(map[string]*DomainWrapper),
		hashserver: hashserver.New(&hashserver.Config{
			Secret: config.Secret,
		}),
		errc:          make(chan error, 10),
		embargo:       true,
		email:         config.Email,
		primaryDomain: config.PrimaryDomain.Name,
		cacheDir:      config.CacheDir,
	}

	addDomain := func(domain *Domain) error {

		if domain.Name == "" {
			return fmt.Errorf("Domain is required")
		}

		if logger.Trace {
			zap.L().Debug(fmt.Sprintf("Adding domain %s", domain.Name))
		}

		s.domains[domain.Name] = &DomainWrapper{
			Domain: domain,
			Server: s,
		}

		return nil
	}

	err := addDomain(config.PrimaryDomain)
	if err != nil {
		return nil, err
	}

	for _, domain := range config.Domains {
		err := addDomain(domain)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (t *Server) Run(ctx context.Context) error {

	if logger.Trace {
		zap.L().Debug("func (t *Server) Run(ctx context.Context) error")
	}

	ctx, cancelCtx := context.WithCancel(ctx)

	liftEmbargo := func() {
		t.mutex.Lock()
		defer t.mutex.Unlock()
		if logger.Trace {
			zap.L().Debug("Lifting embargo")
		}
		t.embargo = false
	}

	defer func() {
		if logger.Trace {
			zap.L().Debug("defer")
		}
		cancelCtx()
		t.stopServer()
		t.hashserver.Shutdown()
		close(t.errc)
	}()

	zap.L().Debug("Processing Domains")

	primaryDomain := t.domains[t.primaryDomain]
	err := primaryDomain.init()
	if err != nil {
		return err
	}

	for _, domain := range t.domains {
		domain.init()
	}

	zap.L().Debug("Processing Domains Completed")

	if primaryDomain.err != nil {
		return fmt.Errorf("Failed to process primary domain; had error %s", primaryDomain.err.Error())
	}

	liftEmbargo()
	t.startServer()

	go func() {
		<-ctx.Done()
		t.errc <- nil
	}()

	if logger.Trace {
		zap.L().Debug("t.wg.Wait()")
	}

	t.wg.Wait()

	return <-t.errc
}

func (t *Server) startServer() {

	if logger.Trace {
		zap.L().Debug("func (t *Server) startServer()")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.embargo {
		if logger.Trace {
			zap.L().Debug("Embargo is in place")
		}
		return
	}

	if logger.Trace {
		zap.L().Debug("Embargo is NOT in place")
	}

	if t.cancel != nil {
		if logger.Trace {
			zap.L().Debug("Server is already running")
		}
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	baseDir := filepath.Join(t.cacheDir, t.primaryDomain)
	httpServer := &http.Server{
		Addr:    ":443",
		Handler: t,
	}

	sendErr := func(err error) {
		if err != nil {
			if err == http.ErrServerClosed {
				return
			}
			t.errc <- err
		}
	}

	go func() {
		zap.L().Debug("Starting ListenAndServeTLS : blocking")
		err := httpServer.ListenAndServeTLS(filepath.Join(baseDir, certPemFileName), filepath.Join(baseDir, keyPemFileName))
		zap.L().Debug("Stopping ListenAndServeTLS : not blocking")
		sendErr(err)
		cancel()
	}()

	go func() {

		if logger.Trace {
			zap.L().Debug("blocking start")
		}

		<-ctx.Done()

		if logger.Trace {
			zap.L().Debug("blocking end")
		}

		zap.L().Debug("Shutting down ListenAndServeTLS")
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelShutdown()
		sendErr(httpServer.Shutdown(ctxShutdown))
		zap.L().Debug("ListenAndServeTLS shut down")
	}()

}

func (t *Server) stopServer() {

	if logger.Trace {
		zap.L().Debug("stopServer()")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.cancel == nil {
		zap.L().Debug("Server is not running")
		return
	}

	t.cancel()
	t.cancel = nil
}

func (t *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	serveHTTP := func() any {

		zap.L().Debug(fmt.Sprintf("Handling %s:%s", r.Method, r.URL.Path))

		switch r.URL.Path {

		case "/getauthrequest":
			return t.hashserver.NewRequest()

		case "/getauthtoken":

			response := &TokenResponse{}

			postBytes, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			authRequest := &AuthRequest{}
			err = json.Unmarshal(postBytes, authRequest)

			if err != nil {
				response.Error = err.Error()
				zap.L().Error(err.Error())
				return response
			}

			token, err := t.hashserver.GetTokenFromRequest(authRequest.AuthRequest)
			if err != nil {
				response.Error = err.Error()
				if logger.Trace {
					zap.L().Debug(err.Error())
				}
				return response
			}

			return token

		case "/getcert":

			response := &CertResponse{}

			authHeader := r.Header.Get("Authorization")
			bearerToken := strings.TrimPrefix(authHeader, prefixBearer)
			if bearerToken == "" {
				response.Error = "bearerToken not found"
				zap.L().Debug("bearerToken not found")
				return response
			}

			domainParam := r.URL.Query().Get("domain")

			zap.L().Debug(fmt.Sprintf("request for domain %s", domainParam))

			if domainParam == "" {
				response.Error = "domain is required"
				zap.L().Debug("domain missing from request")
				return response
			}

			domain := t.domains[domainParam]
			if domain == nil {
				response.Error = "domain not found"
				zap.L().Debug("domain not found")
				return response
			}

			err := t.hashserver.ValidateToken(bearerToken)
			if err != nil {
				response.Error = err.Error()
				if logger.Trace {
					zap.L().Debug(err.Error())
				}
				return response
			}

			cr, err := domain.get()

			if cr != nil {
				response.CR = cr
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("domain %s has non nil CR", domain.Name))
				}
			} else {
				zap.L().Debug(fmt.Sprintf("domain %s has nil CR", domain.Name))
			}

			if err != nil {
				response.Error = err.Error()
				zap.L().Debug(fmt.Sprintf("domain %s has error %s", domain.Name, err.Error()))
			}

			return response
		}

		message := "Valid calls are\n"
		message += fmt.Sprintf("GET https:/%s/getauthrequest\n", r.Host)
		message += fmt.Sprintf("POST https:/%s/getauthtoken\n", r.Host)
		message += fmt.Sprintf("GET https:/%s/getcert?domain=example.com\n", r.Host)

		return &SimpleMessage{
			Message: message,
		}

	}

	w.Header().Set("Content-Type", "application/json")
	response := serveHTTP()
	b, _ := json.Marshal(response)
	fmt.Fprintf(w, string(b))

	if logger.Wire {
		httpDebug := types.NewHTTPDebug(r, b)
		b, _ := json.Marshal(httpDebug)
		zap.L().Debug(fmt.Sprintf("HTTPDebug->%s", string(b)))
	}

}
