package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeConnector struct {
	openCount atomic.Int64
}

func (f *fakeConnector) Connect(ctx context.Context) (driver.Conn, error) {
	f.openCount.Add(1)
	return &fakeConn{}, nil
}

func (f *fakeConnector) Driver() driver.Driver {
	return &fakeDriver{}
}

type fakeDriver struct{}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{}, nil
}

type fakeConn struct{}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeStmt{}, nil
}

func (c *fakeConn) Close() error { return nil }

func (c *fakeConn) Begin() (driver.Tx, error) {
	return &fakeTx{}, nil
}

func (c *fakeConn) Ping(ctx context.Context) error { return nil }

type fakeStmt struct{}

func (s *fakeStmt) Close() error                                    { return nil }
func (s *fakeStmt) NumInput() int                                   { return 0 }
func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) { return nil, nil }
func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error)  { return nil, nil }

type fakeTx struct{}

func (t *fakeTx) Commit() error   { return nil }
func (t *fakeTx) Rollback() error { return nil }

type trackingDBOpener struct {
	callCount atomic.Int64
}

func (o *trackingDBOpener) OpenDB(connector driver.Connector) *sql.DB {
	o.callCount.Add(1)
	return sql.OpenDB(connector)
}

func TestMakeRegistryKey_SameInputsSameKey(t *testing.T) {
	key1 := MakeRegistryKey("host1", "1433", "mydb", "user", "pass")
	key2 := MakeRegistryKey("host1", "1433", "mydb", "user", "pass")

	if key1 != key2 {
		t.Errorf("expected identical keys for same inputs, got %v and %v", key1, key2)
	}
}

func TestMakeRegistryKey_DifferentInputsDifferentKey(t *testing.T) {
	key1 := MakeRegistryKey("host1", "1433", "mydb", "user", "pass")
	key2 := MakeRegistryKey("host1", "1433", "otherdb", "user", "pass")
	key3 := MakeRegistryKey("host1", "1433", "mydb", "user", "differentpass")
	key4 := MakeRegistryKey("host2", "1433", "mydb", "user", "pass")

	if key1 == key2 {
		t.Error("expected different keys for different databases")
	}
	if key1 == key3 {
		t.Error("expected different keys for different credentials")
	}
	if key1 == key4 {
		t.Error("expected different keys for different hosts")
	}
}

func TestDefaultPoolConfig(t *testing.T) {
	config := DefaultPoolConfig()

	if config.MaxOpenConnections != 0 {
		t.Errorf("expected MaxOpenConnections=0, got %d", config.MaxOpenConnections)
	}
	if config.MaxIdleConnections != 2 {
		t.Errorf("expected MaxIdleConnections=2, got %d", config.MaxIdleConnections)
	}
	if config.ConnectionMaxLifetime != 0 {
		t.Errorf("expected ConnectionMaxLifetime=0, got %v", config.ConnectionMaxLifetime)
	}
	if config.ConnectionMaxIdleTime != 0 {
		t.Errorf("expected ConnectionMaxIdleTime=0, got %v", config.ConnectionMaxIdleTime)
	}
}

func TestConnectionRegistry_GetReturnsSameInstance(t *testing.T) {
	opener := &trackingDBOpener{}
	registry := NewConnectionRegistryWithOpener(DefaultPoolConfig(), opener)
	defer registry.Close()

	connector := &fakeConnector{}
	key := MakeRegistryKey("localhost", "1433", "testdb", "sa", "password")

	db1, err := registry.Get(key, connector, 5*time.Second)
	if err != nil {
		t.Fatalf("first Get failed: %v", err)
	}

	db2, err := registry.Get(key, connector, 5*time.Second)
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}

	if db1 != db2 {
		t.Error("expected same *sql.DB instance for same key, got different instances")
	}

	if opener.callCount.Load() != 1 {
		t.Errorf("expected OpenDB called once, got %d", opener.callCount.Load())
	}
}

func TestConnectionRegistry_DifferentKeysReturnDifferentInstances(t *testing.T) {
	opener := &trackingDBOpener{}
	registry := NewConnectionRegistryWithOpener(DefaultPoolConfig(), opener)
	defer registry.Close()

	connector := &fakeConnector{}
	key1 := MakeRegistryKey("localhost", "1433", "db1", "sa", "password")
	key2 := MakeRegistryKey("localhost", "1433", "db2", "sa", "password")

	db1, err := registry.Get(key1, connector, 5*time.Second)
	if err != nil {
		t.Fatalf("first Get failed: %v", err)
	}

	db2, err := registry.Get(key2, connector, 5*time.Second)
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}

	if db1 == db2 {
		t.Error("expected different *sql.DB instances for different keys")
	}

	if opener.callCount.Load() != 2 {
		t.Errorf("expected OpenDB called twice, got %d", opener.callCount.Load())
	}
}

func TestConnectionRegistry_Size(t *testing.T) {
	opener := &trackingDBOpener{}
	registry := NewConnectionRegistryWithOpener(DefaultPoolConfig(), opener)
	defer registry.Close()

	connector := &fakeConnector{}

	if registry.Size() != 0 {
		t.Errorf("expected size 0, got %d", registry.Size())
	}

	registry.Get(MakeRegistryKey("h", "1433", "db1", "u", "p"), connector, 5*time.Second)
	if registry.Size() != 1 {
		t.Errorf("expected size 1, got %d", registry.Size())
	}

	registry.Get(MakeRegistryKey("h", "1433", "db2", "u", "p"), connector, 5*time.Second)
	if registry.Size() != 2 {
		t.Errorf("expected size 2, got %d", registry.Size())
	}

	// Same key should not increase size
	registry.Get(MakeRegistryKey("h", "1433", "db1", "u", "p"), connector, 5*time.Second)
	if registry.Size() != 2 {
		t.Errorf("expected size still 2, got %d", registry.Size())
	}
}

func TestConnectionRegistry_Close(t *testing.T) {
	opener := &trackingDBOpener{}
	registry := NewConnectionRegistryWithOpener(DefaultPoolConfig(), opener)

	connector := &fakeConnector{}
	registry.Get(MakeRegistryKey("h", "1433", "db1", "u", "p"), connector, 5*time.Second)
	registry.Get(MakeRegistryKey("h", "1433", "db2", "u", "p"), connector, 5*time.Second)

	if registry.Size() != 2 {
		t.Fatalf("expected size 2 before close, got %d", registry.Size())
	}

	err := registry.Close()
	if err != nil {
		t.Errorf("expected no error on close, got %v", err)
	}

	if registry.Size() != 0 {
		t.Errorf("expected size 0 after close, got %d", registry.Size())
	}
}

func TestConnectionRegistry_PoolConfigApplied(t *testing.T) {
	config := PoolConfig{
		MaxOpenConnections:    5,
		MaxIdleConnections:    3,
		ConnectionMaxLifetime: 10 * time.Minute,
		ConnectionMaxIdleTime: 4 * time.Minute,
	}
	opener := &trackingDBOpener{}
	registry := NewConnectionRegistryWithOpener(config, opener)
	defer registry.Close()

	connector := &fakeConnector{}
	key := MakeRegistryKey("localhost", "1433", "testdb", "sa", "password")

	db, err := registry.Get(key, connector, 5*time.Second)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	stats := db.Stats()
	if stats.MaxOpenConnections != 5 {
		t.Errorf("expected MaxOpenConnections=5, got %d", stats.MaxOpenConnections)
	}

	// Verify config is stored correctly
	storedConfig := registry.Config()
	if storedConfig.MaxOpenConnections != 5 {
		t.Errorf("expected stored MaxOpenConnections=5, got %d", storedConfig.MaxOpenConnections)
	}
	if storedConfig.MaxIdleConnections != 3 {
		t.Errorf("expected stored MaxIdleConnections=3, got %d", storedConfig.MaxIdleConnections)
	}
	if storedConfig.ConnectionMaxLifetime != 10*time.Minute {
		t.Errorf("expected stored ConnectionMaxLifetime=10m, got %v", storedConfig.ConnectionMaxLifetime)
	}
	if storedConfig.ConnectionMaxIdleTime != 4*time.Minute {
		t.Errorf("expected stored ConnectionMaxIdleTime=4m, got %v", storedConfig.ConnectionMaxIdleTime)
	}
}

func TestConnectionRegistry_ConcurrentAccess(t *testing.T) {
	opener := &trackingDBOpener{}
	registry := NewConnectionRegistryWithOpener(DefaultPoolConfig(), opener)
	defer registry.Close()

	connector := &fakeConnector{}
	key := MakeRegistryKey("localhost", "1433", "testdb", "sa", "password")

	var wg sync.WaitGroup
	results := make([]*sql.DB, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			db, err := registry.Get(key, connector, 5*time.Second)
			if err != nil {
				t.Errorf("goroutine %d: Get failed: %v", idx, err)
				return
			}
			results[idx] = db
		}(i)
	}

	wg.Wait()

	// All goroutines should have received the same *sql.DB instance
	first := results[0]
	for i, db := range results {
		if db != first {
			t.Errorf("goroutine %d got different *sql.DB instance", i)
		}
	}

	// OpenDB should have been called exactly once since the lock is held during creation
	if opener.callCount.Load() != 1 {
		t.Errorf("expected OpenDB called once, got %d", opener.callCount.Load())
	}
}

func TestConnectionRegistry_GetFailsOnConnectionError(t *testing.T) {
	opener := &trackingDBOpener{}
	registry := NewConnectionRegistryWithOpener(DefaultPoolConfig(), opener)
	defer registry.Close()

	failConnector := &failingConnector{}
	key := MakeRegistryKey("unreachable", "1433", "testdb", "sa", "password")

	_, err := registry.Get(key, failConnector, 1*time.Second)
	if err == nil {
		t.Error("expected error for unreachable host, got nil")
	}

	if registry.Size() != 0 {
		t.Errorf("expected size 0 after failed connection, got %d", registry.Size())
	}
}

type failingConnector struct{}

func (f *failingConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return nil, &net.OpError{Op: "dial", Net: "tcp", Addr: nil, Err: fmt.Errorf("connection refused")}
}

func (f *failingConnector) Driver() driver.Driver {
	return &fakeDriver{}
}
