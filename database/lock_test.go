// Copyright (c) 2014-2019 Ludovic Fauzet
// Licensed under the MIT license

package database

import (
	"testing"

	"github.com/rafaeljusto/redigomock"
)

func TestAcquireLockInvalidName(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	_, err := r.AcquireLock("")
	if err != ErrInvalidLockName {
		t.Fatalf("Expected ErrInvalidLockName, got %v", err)
	}
}

func TestAcquireLockSuccess(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("SET", "LOCK_test", redigomock.NewAnyData(), "NX", "PX", "5000").Expect("OK")
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	l, err := r.AcquireLock("test")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if l == nil {
		t.Fatal("Expected non-nil lock")
	}
	if !l.Held() {
		t.Fatal("Expected held to be true")
	}
	l.Release()
}

func TestAcquireLockAlreadyLocked(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("SET", "LOCK_test", redigomock.NewAnyData(), "NX", "PX", "5000").ExpectError(nil)
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	_, err := r.AcquireLock("test")
	_ = err
}

func TestLockReleaseNotHeldExtra(t *testing.T) {
	l := &Lock{held: false}
	l.Release()
}

func TestLockHeldExtra(t *testing.T) {
	l := &Lock{held: true}
	if !l.Held() {
		t.Fatal("Expected true")
	}
	l.held = false
	if l.Held() {
		t.Fatal("Expected false")
	}
}

func TestLockIsValidExtra(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("GET", "LOCK_test").Expect("value")
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	l := &Lock{
		redis: r,
		name:  "LOCK_test",
		value: "value",
		held:  true,
	}
	valid, err := l.isValid()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !valid {
		t.Fatal("Expected valid")
	}
}

func TestLockIsValidWrongValueExtra(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("GET", "LOCK_test").Expect("wrong")
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	l := &Lock{
		redis: r,
		name:  "LOCK_test",
		value: "value",
		held:  true,
	}
	valid, _ := l.isValid()
	if valid {
		t.Fatal("Expected invalid")
	}
}

func TestLockReleaseWithMatchingValueExtra(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("GET", "LOCK_test").Expect("value")
	mock.Command("DEL", "LOCK_test").Expect(int64(1))
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	l := &Lock{
		redis: r,
		name:  "LOCK_test",
		value: "value",
		held:  true,
	}
	l.Release()
	if l.Held() {
		t.Fatal("Expected false after release")
	}
}

func TestErrInvalidLockNameExtra(t *testing.T) {
	if ErrInvalidLockName == nil {
		t.Fatal("Expected non-nil")
	}
}

func TestErrAlreadyLockedExtra(t *testing.T) {
	if ErrAlreadyLocked == nil {
		t.Fatal("Expected non-nil")
	}
}
