package audit

import (
	"time"

	"update-gateway-domain/internal/store"
)

type Entry struct {
	Time            string `json:"time"`
	AccountID       string `json:"account_id,omitempty"`
	Action          string `json:"action"`
	GatewayID       string `json:"gateway_id"`
	Domain          string `json:"domain,omitempty"`
	CertBefore      string `json:"cert_before,omitempty"`
	CertAfter       string `json:"cert_after,omitempty"`
	Protocol        string `json:"protocol,omitempty"`
	ProtocolBefore  string `json:"protocol_before,omitempty"`
	ProtocolAfter   string `json:"protocol_after,omitempty"`
	TLSVersion      string `json:"tls_version"`
	TLSBefore       string `json:"tls_before,omitempty"`
	TLSAfter        string `json:"tls_after,omitempty"`
	MustHTTPS       bool   `json:"must_https"`
	MustHTTPSBefore *bool  `json:"must_https_before,omitempty"`
	MustHTTPSAfter  *bool  `json:"must_https_after,omitempty"`
	HTTP2           string `json:"http2"`
	HTTP2Before     string `json:"http2_before,omitempty"`
	HTTP2After      string `json:"http2_after,omitempty"`
	DryRun          bool   `json:"dry_run"`
	Result          string `json:"result"`
	Message         string `json:"message,omitempty"`
}

type Logger struct {
	store *store.Store
}

func NewLogger(store *store.Store) *Logger {
	return &Logger{store: store}
}

func (l *Logger) Write(entry Entry) error {
	if l.store == nil {
		return nil
	}
	log := &store.AuditLog{
		AccountID:       entry.AccountID,
		Action:          entry.Action,
		GatewayID:       entry.GatewayID,
		Domain:          entry.Domain,
		CertBefore:      entry.CertBefore,
		CertAfter:       entry.CertAfter,
		Protocol:        entry.Protocol,
		ProtocolBefore:  entry.ProtocolBefore,
		ProtocolAfter:   entry.ProtocolAfter,
		TLSVersion:      entry.TLSVersion,
		TLSBefore:       entry.TLSBefore,
		TLSAfter:        entry.TLSAfter,
		MustHTTPS:       entry.MustHTTPS,
		MustHTTPSBefore: entry.MustHTTPSBefore,
		MustHTTPSAfter:  entry.MustHTTPSAfter,
		HTTP2:           entry.HTTP2,
		HTTP2Before:     entry.HTTP2Before,
		HTTP2After:      entry.HTTP2After,
		DryRun:          entry.DryRun,
		Result:          entry.Result,
		Message:         entry.Message,
	}
	return l.store.WriteAuditLog(log)
}

func (l *Logger) ReadLatest(limit int) ([]Entry, error) {
	if l.store == nil {
		return []Entry{}, nil
	}
	if limit <= 0 {
		limit = 50
	}
	logs, _, err := l.store.QueryAuditLogs("", "", "", "", 1, limit)
	if err != nil {
		return nil, err
	}
	return toEntries(logs), nil
}

func (l *Logger) ReadAll() ([]Entry, error) {
	if l.store == nil {
		return []Entry{}, nil
	}
	logs, _, err := l.store.QueryAuditLogs("", "", "", "", 1, 10000)
	if err != nil {
		return nil, err
	}
	return toEntries(logs), nil
}

func (l *Logger) Query(accountID, gatewayID, action, result string, page, pageSize int) ([]Entry, int64, error) {
	if l.store == nil {
		return []Entry{}, 0, nil
	}
	logs, total, err := l.store.QueryAuditLogs(accountID, gatewayID, action, result, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return toEntries(logs), total, nil
}

func toEntries(logs []store.AuditLog) []Entry {
	entries := make([]Entry, 0, len(logs))
	for _, log := range logs {
		entries = append(entries, Entry{
			Time:            log.CreatedAt.Format(time.RFC3339),
			AccountID:       log.AccountID,
			Action:          log.Action,
			GatewayID:       log.GatewayID,
			Domain:          log.Domain,
			CertBefore:      log.CertBefore,
			CertAfter:       log.CertAfter,
			Protocol:        log.Protocol,
			ProtocolBefore:  log.ProtocolBefore,
			ProtocolAfter:   log.ProtocolAfter,
			TLSVersion:      log.TLSVersion,
			TLSBefore:       log.TLSBefore,
			TLSAfter:        log.TLSAfter,
			MustHTTPS:       log.MustHTTPS,
			MustHTTPSBefore: log.MustHTTPSBefore,
			MustHTTPSAfter:  log.MustHTTPSAfter,
			HTTP2:           log.HTTP2,
			HTTP2Before:     log.HTTP2Before,
			HTTP2After:      log.HTTP2After,
			DryRun:          log.DryRun,
			Result:          log.Result,
			Message:         log.Message,
		})
	}
	return entries
}
