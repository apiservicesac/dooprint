package testutil

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
)

// formatMessage formats optional message arguments or returns a default fallback.
func formatMessage(defaultMsg string, msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 {
		return defaultMsg
	}
	if format, ok := msgAndArgs[0].(string); ok {
		if len(msgAndArgs) > 1 {
			return fmt.Sprintf(format, msgAndArgs[1:]...)
		}
		return format
	}
	return fmt.Sprint(msgAndArgs...)
}

// ExpectedTrue asserts that the given condition is true.
func ExpectedTrue(t testing.TB, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !condition {
		msg := formatMessage("expected condition to be true, got false", msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

// ExpectedFalse asserts that the given condition is false.
func ExpectedFalse(t testing.TB, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if condition {
		msg := formatMessage("expected condition to be false, got true", msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

// ExpectedEqual asserts that actual equals expected.
func ExpectedEqual[T comparable](t testing.TB, actual, expected T, msgAndArgs ...interface{}) {
	t.Helper()
	if actual != expected {
		msg := formatMessage(fmt.Sprintf("expected %v, got %v", expected, actual), msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

// ExpectedNotEqual asserts that actual is not equal to expected.
func ExpectedNotEqual[T comparable](t testing.TB, actual, expected T, msgAndArgs ...interface{}) {
	t.Helper()
	if actual == expected {
		msg := formatMessage(fmt.Sprintf("expected value not to equal %v", expected), msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

// ExpectedBytesEqual asserts that actual byte slice equals expected byte slice.
func ExpectedBytesEqual(t testing.TB, actual, expected []byte, msgAndArgs ...interface{}) {
	t.Helper()
	if !bytes.Equal(actual, expected) {
		msg := formatMessage(fmt.Sprintf("expected bytes %v, got %v", expected, actual), msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

// ExpectedNoError asserts that err is nil.
func ExpectedNoError(t testing.TB, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		msg := formatMessage(fmt.Sprintf("expected no error, got: %v", err), msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

// ExpectedError asserts that err is not nil.
func ExpectedError(t testing.TB, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		msg := formatMessage("expected error, got nil", msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

// ExpectedErrorContains asserts that err is non-nil and contains the specified substring.
func ExpectedErrorContains(t testing.TB, err error, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		msg := formatMessage(fmt.Sprintf("expected error containing %q, got nil", substr), msgAndArgs...)
		t.Fatal(msg)
		return
	}
	if !strings.Contains(err.Error(), substr) {
		msg := formatMessage(fmt.Sprintf("expected error %q to contain %q", err.Error(), substr), msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

// ExpectedContains asserts that str contains substr.
func ExpectedContains(t testing.TB, str, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	if !strings.Contains(str, substr) {
		msg := formatMessage(fmt.Sprintf("expected string %q to contain %q", str, substr), msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	default:
		return false
	}
}

func ExpectedNil(t testing.TB, val interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !isNil(val) {
		t.Fatal(formatMessage(fmt.Sprintf("expected nil, got %v", val), msgAndArgs...))
	}
}

func ExpectedNotNil(t testing.TB, val interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if isNil(val) {
		t.Fatal(formatMessage("expected non-nil value, got nil", msgAndArgs...))
	}
}

// ExpectedLen asserts that slice has the expected length.
func ExpectedLen[T any](t testing.TB, slice []T, expectedLen int, msgAndArgs ...interface{}) {
	t.Helper()
	if len(slice) != expectedLen {
		msg := formatMessage(fmt.Sprintf("expected slice length %d, got %d", expectedLen, len(slice)), msgAndArgs...)
		t.Fatal(msg)
		return
	}
}

// StartMockTCPServer starts a local TCP listener on 127.0.0.1:<port> (or 0 for ephemeral)
// and registers automatic cleanup with t.Cleanup.
// For fixed ports (e.g. standard LANPort 9100), a short retry loop with backoff is used
// specifically to handle transient local port reuse while previous sockets exit TIME_WAIT.
func StartMockTCPServer(t testing.TB, onConnect ...func(conn net.Conn)) (net.Listener, int, error) {
	t.Helper()
	var ln net.Listener
	var err error
	port := 9100

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	for range 10 {
		ln, err = net.Listen("tcp", addr)
		if err == nil {
			break
		}
		if port == 0 {
			break
		}
		// If fixed port is in TIME_WAIT from a previous test, pause briefly to let the OS release it
		time.Sleep(30 * time.Millisecond)
	}

	if err != nil {
		return nil, 0, err
	}

	actualPort := ln.Addr().(*net.TCPAddr).Port
	t.Cleanup(func() {
		_ = ln.Close()
	})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			if len(onConnect) > 0 && onConnect[0] != nil {
				go onConnect[0](conn)
			} else {
				_ = conn.Close()
			}
		}
	}()

	return ln, actualPort, nil
}

// GetFreePort finds an available ephemeral TCP port and returns it.
func GetFreePort(t testing.TB) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	ExpectedNoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}
