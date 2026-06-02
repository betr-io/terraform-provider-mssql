package sql

import (
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"time"
)

type PoolConfig struct {
	MaxOpenConnections    int
	MaxIdleConnections    int
	ConnectionMaxLifetime time.Duration
	ConnectionMaxIdleTime time.Duration
}

func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConnections:    0,
		MaxIdleConnections:    2,
		ConnectionMaxLifetime: 0,
		ConnectionMaxIdleTime: 0,
	}
}

type RegistryKey struct {
	Host           string
	Port           string
	Database       string
	CredentialHash string
}

type ConnectionRegistry struct {
	mu     sync.Mutex
	pools  map[RegistryKey]*sql.DB
	config PoolConfig
	opener DBOpener
}

type DBOpener interface {
	OpenDB(connector driver.Connector) *sql.DB
}

type defaultDBOpener struct{}

func (d *defaultDBOpener) OpenDB(connector driver.Connector) *sql.DB {
	return sql.OpenDB(connector)
}

func NewConnectionRegistry(config PoolConfig) *ConnectionRegistry {
	return &ConnectionRegistry{
		pools:  make(map[RegistryKey]*sql.DB),
		config: config,
		opener: &defaultDBOpener{},
	}
}

func NewConnectionRegistryWithOpener(config PoolConfig, opener DBOpener) *ConnectionRegistry {
	return &ConnectionRegistry{
		pools:  make(map[RegistryKey]*sql.DB),
		config: config,
		opener: opener,
	}
}

// Get returns a cached *sql.DB for the given key and connector, or creates and caches a new one.
// The lock is held during connection creation to prevent duplicate connections for the same key.
func (r *ConnectionRegistry) Get(key RegistryKey, connector driver.Connector, timeout time.Duration) (*sql.DB, error) {
	r.mu.Lock()
	if db, ok := r.pools[key]; ok {
		r.mu.Unlock()
		return db, nil
	}

	db, err := r.openWithRetry(connector, timeout)
	if err != nil {
		r.mu.Unlock()
		return nil, err
	}

	r.applyPoolConfig(db)
	r.pools[key] = db
	r.mu.Unlock()
	return db, nil
}

// Close closes all pooled connections and clears the registry.
func (r *ConnectionRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for key, db := range r.pools {
		if err := db.Close(); err != nil {
			lastErr = err
		}
		delete(r.pools, key)
	}
	return lastErr
}

func (r *ConnectionRegistry) Size() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.pools)
}

func (r *ConnectionRegistry) Config() PoolConfig {
	return r.config
}

func (r *ConnectionRegistry) applyPoolConfig(db *sql.DB) {
	db.SetMaxOpenConns(r.config.MaxOpenConnections)
	db.SetMaxIdleConns(r.config.MaxIdleConnections)
	db.SetConnMaxLifetime(r.config.ConnectionMaxLifetime)
	db.SetConnMaxIdleTime(r.config.ConnectionMaxIdleTime)
}

func (r *ConnectionRegistry) openWithRetry(connector driver.Connector, timeout time.Duration) (*sql.DB, error) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	timeoutExceeded := time.After(timeout)
	for {
		select {
		case <-timeoutExceeded:
			return nil, fmt.Errorf("db connection failed after %s timeout", timeout)

		case <-ticker.C:
			db := r.opener.OpenDB(connector)
			if err := db.Ping(); err == nil {
				return db, nil
			} else {
				db.Close()
				errStr := err.Error()
				if strings.Contains(strings.ToLower(errStr), "login failed") ||
					strings.Contains(strings.ToLower(errStr), "login error") ||
					strings.Contains(errStr, "error retrieving access token") ||
					strings.Contains(errStr, "AuthenticationFailedError") ||
					strings.Contains(errStr, "credential") ||
					strings.Contains(errStr, "request failed") {
					return nil, err
				}
			}
		}
	}
}

func MakeRegistryKey(host, port, database, username, password string) RegistryKey {
	h := sha256.New()
	h.Write([]byte(username))
	h.Write([]byte(":"))
	h.Write([]byte(password))
	credHash := fmt.Sprintf("%x", h.Sum(nil))

	return RegistryKey{
		Host:           host,
		Port:           port,
		Database:       database,
		CredentialHash: credHash,
	}
}
