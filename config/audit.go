package config

import (
	"fmt"
	"strings"

	"github.com/hivepaas/shellwing-server/pkg/auditsink"
)

// The sinks a copy of each audit record can go to.
const (
	// AuditSinkNone keeps the records only in this server's database, which is where they have always been.
	AuditSinkNone = "none"
	// AuditSinkSyslog sends each record to a syslog collector as well.
	AuditSinkSyslog = "syslog"
)

// AuditSinks are the names a configuration may give.
var AuditSinks = []string{AuditSinkNone, AuditSinkSyslog}

// Audit says where a copy of each recorded action goes, besides this server's own database.
//
// The records are evidence about this server, kept by this server: whoever reaches the database can edit or delete
// them. A copy somewhere else is what makes them worth showing to anybody. See pkg/auditsink for what that does and
// does not defend against.
//
// Point it at a collector on this machine - rsyslog, syslog-ng, vector, fluent-bit - and let that forward to wherever
// the records are kept. The collector is what has the disk queue and the credentials for the far end.
type Audit struct {
	// Sink is one of AuditSinks.
	Sink string `toml:"sink" env:"SHW_AUDIT_SINK" default:"none"`
	// Required refuses an action whose record cannot reach the sink.
	//
	// On by default, because configuring a sink is saying that is where the records have to be, and because it is the
	// stance service/auditservice already takes about the database: an action that cannot be recorded does not
	// happen. Turn it off to prefer the server staying up over the records being complete, and read the log: a
	// failure then says what was lost.
	Required bool `toml:"required" env:"SHW_AUDIT_REQUIRED" default:"true"`

	// SyslogAddress is the collector's host:port. 601 is syslog over TCP, 6514 over TLS.
	SyslogAddress string `toml:"syslog_address" env:"SHW_AUDIT_SYSLOG_ADDRESS"`
	// SyslogNetwork is "tls" or "tcp".
	//
	// TLS by default, so that a collector named by an address that is not this machine is reached over one. A
	// collector on this machine is reached over "tcp", which has to be said rather than assumed - the cost of
	// guessing it the other way is audit records in the clear across somebody's network.
	SyslogNetwork string `toml:"syslog_network" env:"SHW_AUDIT_SYSLOG_NETWORK" default:"tls"`
	// SyslogAppName is the APP-NAME field, how a collector tells this server's records from everything else's.
	SyslogAppName string `toml:"syslog_app_name" env:"SHW_AUDIT_SYSLOG_APP_NAME" default:"shellwing"`
	// SyslogFacility is 0 to 23. Zero is local0, which is what an application is meant to use.
	SyslogFacility int `toml:"syslog_facility" env:"SHW_AUDIT_SYSLOG_FACILITY"`
	// SyslogHostname is the HOSTNAME field. Empty, this machine's own name.
	SyslogHostname string `toml:"syslog_hostname" env:"SHW_AUDIT_SYSLOG_HOSTNAME"`
	// SyslogCACertFile is a PEM to trust instead of the system store.
	SyslogCACertFile string `toml:"syslog_ca_cert_file" env:"SHW_AUDIT_SYSLOG_CA_CERT_FILE"`
	// SyslogServerName is the name to verify the collector's certificate against, for an address that is an IP.
	SyslogServerName string `toml:"syslog_server_name" env:"SHW_AUDIT_SYSLOG_SERVER_NAME"`
}

// validate refuses a sink this server does not have, at load.
//
// What the sink needs to reach anything is not checked here: that belongs where it is built, which is the one place
// that knows, and it names the field.
func (a *Audit) validate() error {
	if a.Sink == "" {
		a.Sink = AuditSinkNone
	}
	for _, sink := range AuditSinks {
		if sink == a.Sink {
			return nil
		}
	}
	return fmt.Errorf("%w: SHW_AUDIT_SINK %q - one of %s",
		ErrAuditSinkUnknown, a.Sink, strings.Join(AuditSinks, ", "))
}

// Syslog is the sink's options, as pkg/auditsink wants them.
func (a *Audit) Syslog() auditsink.SyslogOptions {
	return auditsink.SyslogOptions{
		Address:    a.SyslogAddress,
		Network:    a.SyslogNetwork,
		AppName:    a.SyslogAppName,
		Facility:   a.SyslogFacility,
		Hostname:   a.SyslogHostname,
		CACertFile: a.SyslogCACertFile,
		ServerName: a.SyslogServerName,
	}
}
