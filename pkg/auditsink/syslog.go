package auditsink

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// NetworkTCP is a plain connection, for a collector on this machine or across a network already private.
	NetworkTCP = "tcp"
	// NetworkTLS is a connection this server verifies, for a collector reached across anything else.
	NetworkTLS = "tls"

	// FacilityMax is the highest syslog facility. local0 is 16, and local0..local7 are what an application is
	// supposed to use.
	FacilityMax = 23
	// defaultFacility is local0.
	defaultFacility = 16
	// severityInfo is what a recorded action is. A refused one is still an action that happened, so the severity does
	// not move with the result: whoever reads these filters on the record's own fields, not on a number that only
	// coarsely repeats one of them.
	severityInfo = 6

	// defaultTimeout bounds one delivery. It is short because a request is waiting on it.
	defaultTimeout = 5 * time.Second
	// defaultMaxPayload bounds one record. Receivers truncate what is too long - rsyslog at 8 KiB unless told
	// otherwise - and a truncated record is invalid JSON, which is worse than a record that says it was too long.
	defaultMaxPayload = 6 << 10
	// msgIDMax is RFC 5424's limit for MSGID.
	msgIDMax = 32
	// decimalBase is what the priority is written in.
	decimalBase = 10
	// probeTimeout is how long alive waits to hear that the connection has ended. An end that has already arrived
	// reads immediately, so this only has to be longer than nothing.
	probeTimeout = time.Millisecond
)

// Networks are the ones a configuration may name.
//
// UDP is not among them on purpose. It drops records without saying so, which is the one failure this whole package is
// meant to make impossible: a sink that silently loses the records it was given is indistinguishable from no sink, and
// would be believed.
var Networks = []string{NetworkTCP, NetworkTLS}

// SyslogOptions is what reaching a collector needs.
type SyslogOptions struct {
	// Address is host:port.
	Address string
	// Network is NetworkTCP or NetworkTLS.
	Network string
	// AppName is the APP-NAME field, which is how a collector tells this server's records from everything else's.
	AppName string
	// Facility is 0..FacilityMax. Zero means local0.
	Facility int
	// Hostname is the HOSTNAME field. Empty, the machine's own name is used.
	Hostname string
	// CACertFile is a PEM to trust instead of the system store, for a collector behind a private authority.
	CACertFile string
	// ServerName is the name to verify the collector's certificate against, for an Address that is an IP.
	ServerName string
	// Timeout bounds one delivery. Zero is defaultTimeout.
	Timeout time.Duration
	// MaxPayload bounds one record. Zero is defaultMaxPayload.
	MaxPayload int
}

// syslogSink writes RFC 5424 records over a connection it keeps open.
type syslogSink struct {
	address    string
	network    string
	appName    string
	hostname   string
	procID     string
	priority   int
	timeout    time.Duration
	maxPayload int
	tlsConfig  *tls.Config

	// mu serializes delivery. Records go out one at a time on one connection, which is what keeps them in order and
	// keeps two of them from interleaving halfway through a line.
	mu   sync.Mutex
	conn net.Conn
}

// NewSyslog reaches a collector. It does not connect: the first record does, so that a collector started a moment
// after this server is not a reason the server will not start.
func NewSyslog(opts SyslogOptions) (Sink, error) {
	if opts.Address == "" {
		return nil, fmt.Errorf("%w: syslog needs an address", ErrNotConfigured)
	}
	if opts.Network != NetworkTCP && opts.Network != NetworkTLS {
		return nil, fmt.Errorf("%w: syslog network %q - one of %s",
			ErrNotConfigured, opts.Network, strings.Join(Networks, ", "))
	}
	if opts.Facility < 0 || opts.Facility > FacilityMax {
		return nil, fmt.Errorf("%w: syslog facility %d - 0 to %d",
			ErrNotConfigured, opts.Facility, FacilityMax)
	}

	sink := &syslogSink{
		address:    opts.Address,
		network:    opts.Network,
		appName:    nilValue(opts.AppName),
		hostname:   opts.Hostname,
		procID:     strconv.Itoa(os.Getpid()),
		timeout:    opts.Timeout,
		maxPayload: opts.MaxPayload,
	}
	facility := opts.Facility
	if facility == 0 {
		facility = defaultFacility
	}
	sink.priority = facility*8 + severityInfo //nolint:mnd // RFC 5424: PRI is facility times eight plus severity
	if sink.timeout <= 0 {
		sink.timeout = defaultTimeout
	}
	if sink.maxPayload <= 0 {
		sink.maxPayload = defaultMaxPayload
	}
	if sink.hostname == "" {
		if name, err := os.Hostname(); err == nil {
			sink.hostname = name
		}
	}
	sink.hostname = nilValue(sink.hostname)

	if opts.Network == NetworkTLS {
		config, err := syslogTLS(opts)
		if err != nil {
			return nil, err
		}
		sink.tlsConfig = config
	}
	return sink, nil
}

func syslogTLS(opts SyslogOptions) (*tls.Config, error) {
	config := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: opts.ServerName}
	if opts.CACertFile == "" {
		return config, nil
	}
	pem, err := os.ReadFile(opts.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot read the CA certificate %s: %w", ErrNotConfigured, opts.CACertFile, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("%w: %s holds no certificate", ErrNotConfigured, opts.CACertFile)
	}
	config.RootCAs = pool
	return config, nil
}

func (s *syslogSink) Name() string { return "syslog" }

// Send writes one record, reconnecting once if the connection it held was closed.
//
// The retry is what makes a collector's own restart invisible: the first write after it went away fails on a socket
// that looks fine, and without a second attempt every restart of the collector would refuse one action here.
func (s *syslogSink) Send(ctx context.Context, msgID string, payload []byte) error {
	line := s.format(msgID, payload)

	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.write(ctx, line)
	if err == nil {
		return nil
	}
	s.drop()
	if retryErr := s.write(ctx, line); retryErr != nil {
		return fmt.Errorf("%w: %s at %s: %w", ErrSinkUnavailable, s.network, s.address, retryErr)
	}
	return nil
}

// alive reports whether the connection is still open, by asking whether its end has arrived.
//
// This is not a nicety. A write to a connection the far end has closed succeeds here: the kernel takes the bytes and
// only a later write finds out, by which time that record is gone and nothing said so. Syslog has no acknowledgement
// of any kind - not over TCP, not over TLS - so there is nothing to wait for and no way to know a collector received
// anything. The one question that can be asked is whether the end has arrived, and a read with a deadline a moment
// away answers it: an ended connection reads as EOF at once, and one still open reads as a timeout, having consumed
// nothing. That turns a collector's restart from a record lost in silence into a reconnection.
//
// It leaves a window - the collector can stop between this and the write - which no amount of care closes, because
// the protocol has nothing to acknowledge with. That window is why a collector belongs on this machine, with its own
// queue, rather than at the far end of a network.
//
// Nothing a collector sends is expected, so bytes are a reason to start again too: whatever is at the other end is
// not speaking syslog.
func (s *syslogSink) alive() bool {
	if s.conn == nil {
		return false
	}
	if err := s.conn.SetReadDeadline(time.Now().Add(probeTimeout)); err != nil {
		return false
	}
	var one [1]byte
	_, err := s.conn.Read(one[:])
	var timeout net.Error
	return errors.As(err, &timeout) && timeout.Timeout()
}

// write delivers the line on the connection, opening one if there is none.
func (s *syslogSink) write(ctx context.Context, line []byte) error {
	if !s.alive() {
		s.drop()
	}
	if s.conn == nil {
		conn, err := s.dial(ctx)
		if err != nil {
			return err
		}
		s.conn = conn
	}
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) > s.timeout {
		deadline = time.Now().Add(s.timeout)
	}
	if err := s.conn.SetWriteDeadline(deadline); err != nil {
		return err //nolint:wrapcheck // the caller names the sink
	}
	_, err := s.conn.Write(line)
	return err //nolint:wrapcheck // the caller names the sink
}

func (s *syslogSink) dial(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: s.timeout}
	if s.network == NetworkTLS {
		//nolint:wrapcheck // the caller names the sink
		return tls.DialWithDialer(dialer, NetworkTCP, s.address, s.tlsConfig)
	}
	return dialer.DialContext(ctx, NetworkTCP, s.address) //nolint:wrapcheck // the caller names the sink
}

// drop forgets the connection, so the next write opens another.
func (s *syslogSink) drop() {
	if s.conn != nil {
		_ = s.conn.Close()
		s.conn = nil
	}
}

func (s *syslogSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drop()
	return nil
}

// format writes the record as RFC 5424, ending in a newline.
//
// Newline framing rather than RFC 6587's octet counting, because the body is JSON and so can never hold a bare
// newline of its own - which is the only thing octet counting buys - and because every collector takes it.
//
// No byte order mark before the body. The RFC offers one to declare the body UTF-8, and a collector that passes it
// through puts a stray three bytes in front of what the far end has to parse as JSON.
func (s *syslogSink) format(msgID string, payload []byte) []byte {
	body := payload
	if len(body) > s.maxPayload {
		// Said rather than cut. A record truncated in the middle is invalid JSON, which a reader is likely to drop
		// entirely - losing the record this is meant to be a copy of, and saying nothing about why.
		body = tooLong(msgID, len(payload), s.maxPayload)
	}
	line := make([]byte, 0, len(body)+128) //nolint:mnd // room for the header
	line = append(line, '<')
	line = strconv.AppendInt(line, int64(s.priority), decimalBase)
	line = append(line, ">1 "...)
	line = time.Now().UTC().AppendFormat(line, "2006-01-02T15:04:05.000000Z")
	line = append(line, ' ')
	line = append(line, s.hostname...)
	line = append(line, ' ')
	line = append(line, s.appName...)
	line = append(line, ' ')
	line = append(line, s.procID...)
	line = append(line, ' ')
	line = append(line, printableID(msgID)...)
	// No structured data: everything there is to say is in the body, and saying it twice invites the two to disagree.
	line = append(line, " - "...)
	line = append(line, body...)
	return append(line, '\n')
}

// tooLong is what goes out in place of a record that will not fit, as JSON a reader can still parse.
func tooLong(msgID string, size, limit int) []byte {
	return []byte(fmt.Sprintf(
		`{"type":%q,"auditSinkError":"the record is %d bytes, over the %d byte limit, and was not sent"}`,
		msgID, size, limit))
}

// printableID makes a MSGID of a name: RFC 5424 allows printable ASCII up to 32 characters, and no spaces.
func printableID(name string) string {
	if name == "" {
		return "-"
	}
	var out strings.Builder
	for _, r := range name {
		if r < '!' || r > '~' {
			continue
		}
		out.WriteRune(r)
		if out.Len() == msgIDMax {
			break
		}
	}
	return nilValue(out.String())
}

// nilValue is RFC 5424's "-", which is how a field says it has nothing to say.
func nilValue(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
