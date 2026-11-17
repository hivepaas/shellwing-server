package auditsink

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// byteOrderMark is UTF-8's. RFC 5424 offers one in front of the body to declare it UTF-8; this sink leaves it
// out, because a collector that passes it through puts three stray bytes in front of what the far end parses as
// JSON. Written as bytes because gofmt turns the escape into the character, which Go will not have in a source
// file.
var byteOrderMark = []byte{0xEF, 0xBB, 0xBF}

// collector is a listener that reads whole lines, as a syslog collector does.
type collector struct {
	listener net.Listener
	lines    chan string
	mu       sync.Mutex
	conns    []net.Conn
}

func newCollector(t *testing.T, config *tls.Config) *collector {
	t.Helper()
	var listener net.Listener
	var err error
	if config == nil {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
	} else {
		listener, err = tls.Listen("tcp", "127.0.0.1:0", config)
	}
	require.NoError(t, err)

	c := &collector{listener: listener, lines: make(chan string, 16)}
	go c.accept()
	t.Cleanup(func() { _ = listener.Close() })
	return c
}

func (c *collector) accept() {
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			return
		}
		c.mu.Lock()
		c.conns = append(c.conns, conn)
		c.mu.Unlock()
		go func() {
			reader := bufio.NewReader(conn)
			for {
				line, err := reader.ReadString('\n')
				if line != "" {
					c.lines <- strings.TrimSuffix(line, "\n")
				}
				if err != nil {
					return
				}
			}
		}()
	}
}

func (c *collector) address() string { return c.listener.Addr().String() }

// next waits for one line, or fails: a sink that silently sent nothing would otherwise pass.
func (c *collector) next(t *testing.T) string {
	t.Helper()
	select {
	case line := <-c.lines:
		return line
	case <-time.After(3 * time.Second):
		t.Fatal("no record reached the collector")
		return ""
	}
}

// restart drops the connections, as a collector being restarted does.
func (c *collector) restart() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, conn := range c.conns {
		_ = conn.Close()
	}
	c.conns = nil
}

// record is an RFC 5424 line taken apart.
type record struct {
	priority int
	version  string
	stamp    string
	hostname string
	appName  string
	procID   string
	msgID    string
	data     string
	body     string
}

// parse5424 splits a line the way a collector does: eight header fields and then the body.
func parse5424(t *testing.T, line string) record {
	t.Helper()
	require.True(t, strings.HasPrefix(line, "<"), "a record starts with its priority: %q", line)
	end := strings.Index(line, ">")
	require.Positive(t, end)
	priority, err := strconv.Atoi(line[1:end])
	require.NoError(t, err)

	fields := strings.SplitN(line[end+1:], " ", 7)
	require.Len(t, fields, 7, "version, time, host, app, pid, msgid, data and body: %q", line)
	data, body, found := strings.Cut(fields[6], " ")
	require.True(t, found, "structured data then the body: %q", fields[6])

	return record{
		priority: priority,
		version:  fields[0],
		stamp:    fields[1],
		hostname: fields[2],
		appName:  fields[3],
		procID:   fields[4],
		msgID:    fields[5],
		data:     data,
		body:     body,
	}
}

func syslogSinkFor(t *testing.T, c *collector, opts SyslogOptions) Sink {
	t.Helper()
	if opts.Address == "" {
		opts.Address = c.address()
	}
	if opts.Network == "" {
		opts.Network = NetworkTCP
	}
	sink, err := NewSyslog(opts)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sink.Close() })
	return sink
}

func TestItWritesARecordACollectorCanRead(t *testing.T) {
	c := newCollector(t, nil)
	sink := syslogSinkFor(t, c, SyslogOptions{AppName: "shellwing", Hostname: "host-under-test"})
	assert.Equal(t, "syslog", sink.Name())

	payload := []byte(`{"id":"01JC","type":"ssh-cert-issue","actorName":"alice"}`)
	require.NoError(t, sink.Send(context.Background(), "ssh-cert-issue", payload))

	got := parse5424(t, c.next(t))
	assert.Equal(t, 134, got.priority, "local0 and informational")
	assert.Equal(t, "1", got.version)
	assert.Equal(t, "host-under-test", got.hostname)
	assert.Equal(t, "shellwing", got.appName)
	assert.Equal(t, strconv.Itoa(os.Getpid()), got.procID)
	assert.Equal(t, "ssh-cert-issue", got.msgID, "the kind of action is a field of its own")
	assert.Equal(t, "-", got.data)
	assert.JSONEq(t, string(payload), got.body, "the record reaches the collector as it was given")

	stamp, err := time.Parse("2006-01-02T15:04:05.000000Z", got.stamp)
	require.NoError(t, err, "the timestamp is one a collector can parse: %q", got.stamp)
	assert.WithinDuration(t, time.Now().UTC(), stamp, time.Minute)
	assert.False(t, bytes.HasPrefix([]byte(got.body), byteOrderMark), "no byte order mark in front of the JSON")
}

// A collector restarts, and the socket this server holds looks fine until it is written to. Without a second attempt,
// every restart of the collector would refuse one action here.
func TestItReconnectsWhenTheCollectorRestarts(t *testing.T) {
	c := newCollector(t, nil)
	sink := syslogSinkFor(t, c, SyslogOptions{AppName: "shellwing"})
	ctx := context.Background()

	require.NoError(t, sink.Send(ctx, "first", []byte(`{"n":1}`)))
	assert.Equal(t, `{"n":1}`, parse5424(t, c.next(t)).body)

	c.restart()
	// Let the close reach this side, so the next write is the one that finds it.
	time.Sleep(50 * time.Millisecond)

	require.NoError(t, sink.Send(ctx, "second", []byte(`{"n":2}`)), "a restart is not a reason to refuse an action")
	assert.Equal(t, `{"n":2}`, parse5424(t, c.next(t)).body)
}

// The same over TLS, because the probe reads from the connection and a timed-out read is the one thing that could
// leave a TLS connection unusable. Assuming it is fine is not the same as knowing.
func TestItReconnectsOverTLSToo(t *testing.T) {
	certFile, config := selfSigned(t)
	c := newCollector(t, config)
	sink := syslogSinkFor(t, c, SyslogOptions{Network: NetworkTLS, AppName: "shellwing", CACertFile: certFile})
	ctx := context.Background()

	require.NoError(t, sink.Send(ctx, "first", []byte(`{"n":1}`)))
	assert.Equal(t, `{"n":1}`, parse5424(t, c.next(t)).body)

	// Several more on the same connection, so a probe that had broken the TLS state would show.
	for i := range 3 {
		require.NoError(t, sink.Send(ctx, "again", fmt.Appendf(nil, `{"n":%d}`, i)), "record %d", i)
		assert.Equal(t, fmt.Sprintf(`{"n":%d}`, i), parse5424(t, c.next(t)).body)
	}

	c.restart()
	time.Sleep(50 * time.Millisecond)
	require.NoError(t, sink.Send(ctx, "second", []byte(`{"n":9}`)))
	assert.Equal(t, `{"n":9}`, parse5424(t, c.next(t)).body)
}

// A record too long for the collector is reported rather than cut: a line truncated in the middle is invalid JSON,
// which a reader is likely to drop entirely - losing the very copy this exists to make.
func TestItSaysWhenARecordIsTooLong(t *testing.T) {
	c := newCollector(t, nil)
	sink := syslogSinkFor(t, c, SyslogOptions{AppName: "shellwing", MaxPayload: 200})

	huge := `{"id":"01JC","detail":"` + strings.Repeat("x", 500) + `"}`
	require.NoError(t, sink.Send(context.Background(), "setting-update", []byte(huge)))

	body := parse5424(t, c.next(t)).body
	var sent map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &sent), "what goes out is still JSON: %q", body)
	assert.Equal(t, "setting-update", sent["type"], "and still says what kind of action it was")
	assert.Contains(t, sent["auditSinkError"], "over the 200 byte limit")
	assert.NotContains(t, body, strings.Repeat("x", 100), "and carries none of the record it could not send")
}

func TestItRefusesACollectorItCannotBeReached(t *testing.T) {
	sink, err := NewSyslog(SyslogOptions{
		// A port nothing listens on, so this fails rather than hangs.
		Address: "127.0.0.1:1", Network: NetworkTCP, AppName: "shellwing", Timeout: time.Second,
	})
	require.NoError(t, err, "it does not connect until the first record")

	err = sink.Send(context.Background(), "ssh-cert-issue", []byte(`{"n":1}`))
	require.ErrorIs(t, err, ErrSinkUnavailable)
	assert.Contains(t, err.Error(), "127.0.0.1:1", "and says where it tried")
}

// TLS is the mode for a collector reached across anything but this machine, so it is the one worth a real handshake.
func TestItWritesOverTLS(t *testing.T) {
	certFile, config := selfSigned(t)
	c := newCollector(t, config)
	sink := syslogSinkFor(t, c, SyslogOptions{
		Network: NetworkTLS, AppName: "shellwing", CACertFile: certFile,
	})

	require.NoError(t, sink.Send(context.Background(), "ssh-cert-issue", []byte(`{"n":1}`)))
	assert.Equal(t, `{"n":1}`, parse5424(t, c.next(t)).body)
}

func TestItRefusesACollectorWhoseCertificateItCannotVerify(t *testing.T) {
	_, config := selfSigned(t)
	c := newCollector(t, config)
	// No CA file, so the system store decides, and it has never heard of this certificate.
	sink := syslogSinkFor(t, c, SyslogOptions{Network: NetworkTLS, AppName: "shellwing", Timeout: time.Second})

	err := sink.Send(context.Background(), "ssh-cert-issue", []byte(`{"n":1}`))
	require.ErrorIs(t, err, ErrSinkUnavailable, "an unverified collector is not a collector")
}

// Records go out one at a time on one connection. Interleaved halves of two lines would be two broken records, and
// the collector would have no way to tell.
func TestRecordsDoNotInterleave(t *testing.T) {
	c := newCollector(t, nil)
	sink := syslogSinkFor(t, c, SyslogOptions{AppName: "shellwing"})

	const count = 20
	var group sync.WaitGroup
	for i := range count {
		group.Add(1)
		go func() {
			defer group.Done()
			payload := fmt.Appendf(nil, `{"n":%d,"pad":%q}`, i, strings.Repeat("y", 200))
			assert.NoError(t, sink.Send(context.Background(), "ssh-cert-issue", payload))
		}()
	}
	group.Wait()

	seen := map[float64]bool{}
	for range count {
		var sent map[string]any
		body := parse5424(t, c.next(t)).body
		require.NoError(t, json.Unmarshal([]byte(body), &sent), "each line is one whole record: %q", body)
		seen[sent["n"].(float64)] = true //nolint:forcetypeassert // a number this test wrote
	}
	assert.Len(t, seen, count, "every record arrived, once")
}

func TestASyslogSinkNeedsEnoughToReachOne(t *testing.T) {
	for _, c := range []struct {
		name string
		opts SyslogOptions
	}{
		{"no address", SyslogOptions{Network: NetworkTCP}},
		{"no network", SyslogOptions{Address: "127.0.0.1:514"}},
		{"udp, which loses records", SyslogOptions{Address: "127.0.0.1:514", Network: "udp"}},
		{"facility too high", SyslogOptions{Address: "127.0.0.1:514", Network: NetworkTCP, Facility: 24}},
		{"facility negative", SyslogOptions{Address: "127.0.0.1:514", Network: NetworkTCP, Facility: -1}},
	} {
		_, err := NewSyslog(c.opts)
		require.ErrorIs(t, err, ErrNotConfigured, c.name)
	}

	_, err := NewSyslog(SyslogOptions{
		Address: "127.0.0.1:514", Network: NetworkTLS, CACertFile: "/nope/no-such.pem",
	})
	require.ErrorIs(t, err, ErrNotConfigured, "a CA file that is not there is a configuration error")
}

func TestTheMessageIDIsSomethingSyslogAccepts(t *testing.T) {
	assert.Equal(t, "ssh-cert-issue", printableID("ssh-cert-issue"))
	assert.Equal(t, "-", printableID(""), "RFC 5424's nil value")
	assert.Equal(t, "nospaces", printableID("no spaces"), "MSGID has no spaces")
	assert.Equal(t, "ab", printableID("aäb"), "nor anything outside printable ASCII")
	assert.Len(t, printableID(strings.Repeat("a", 64)), msgIDMax, "and is bounded")
}

func TestTheNopSinkKeepsNothing(t *testing.T) {
	sink := Nop()
	assert.Equal(t, "none", sink.Name())
	require.NoError(t, sink.Send(context.Background(), "ssh-cert-issue", []byte(`{"n":1}`)))
	require.NoError(t, sink.Close())
}

// selfSigned makes a certificate for 127.0.0.1 and writes it where a sink can trust it.
func selfSigned(t *testing.T) (certFile string, config *tls.Config) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "collector-under-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certFile = filepath.Join(t.TempDir(), "collector.pem")
	require.NoError(t, os.WriteFile(certFile,
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))

	leaf, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return certFile, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}},
	}
}
