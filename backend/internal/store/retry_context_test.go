package store_test

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"colorpixel/internal/store"
)

func TestOpenRetryCancelsStalledHandshake(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type openResult struct {
		db  *store.DB
		err error
	}
	opened := make(chan openResult, 1)
	url := fmt.Sprintf("postgres://test:test@%s/colorpixel?sslmode=disable&connect_timeout=10", listener.Addr().String())
	go func() {
		db, err := store.OpenRetry(ctx, url, 3)
		opened <- openResult{db: db, err: err}
	}()

	type acceptResult struct {
		conn net.Conn
		err  error
	}
	accepted := make(chan acceptResult, 1)
	go func() {
		conn, err := listener.Accept()
		accepted <- acceptResult{conn: conn, err: err}
	}()

	var conn net.Conn
	select {
	case got := <-accepted:
		if got.err != nil {
			t.Fatal(got.err)
		}
		conn = got.conn
	case got := <-opened:
		if got.db != nil {
			got.db.Close()
		}
		t.Fatalf("OpenRetry returned before the database handshake started: %v", got.err)
	case <-time.After(3 * time.Second):
		t.Fatal("database connection was not attempted")
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := readStartupPacket(conn); err != nil {
		t.Fatalf("read PostgreSQL startup packet: %v", err)
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		t.Fatal(err)
	}

	cancel()
	select {
	case got := <-opened:
		if got.db != nil {
			got.db.Close()
			t.Fatal("OpenRetry returned a database after cancellation")
		}
		if !errors.Is(got.err, context.Canceled) {
			t.Fatalf("OpenRetry error = %v, want context.Canceled", got.err)
		}
	case <-time.After(500 * time.Millisecond):
		_ = conn.Close()
		select {
		case got := <-opened:
			if got.db != nil {
				got.db.Close()
			}
		case <-time.After(2 * time.Second):
			t.Fatal("OpenRetry remained blocked after the peer connection was closed")
		}
		t.Fatal("OpenRetry did not stop the in-flight handshake after cancellation")
	}
}

func readStartupPacket(conn net.Conn) error {
	var header [4]byte
	if _, err := io.ReadFull(conn, header[:]); err != nil {
		return err
	}
	size := binary.BigEndian.Uint32(header[:])
	if size < 8 || size > 64<<10 {
		return fmt.Errorf("invalid startup packet size %d", size)
	}
	_, err := io.CopyN(io.Discard, conn, int64(size)-4)
	return err
}
