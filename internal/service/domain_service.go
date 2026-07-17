package service

import (
	"context"
	"fmt"
	"strings"

	"update-gateway-domain/internal/audit"
	"update-gateway-domain/internal/config"
	"update-gateway-domain/internal/mse"
	"update-gateway-domain/internal/store"
)

var tlsMap = map[string]string{
	"tlsv1.0": "TLS 1.0",
	"tlsv1.2": "TLS 1.2",
	"tlsv1.3": "TLS 1.3",
}

type DomainService struct {
	clientManager *mse.ClientManager
	cfg           *config.Config
	auditLogger   *audit.Logger
	store         *store.Store
}

type DomainPage struct {
	Total   int                 `json:"total"`
	Page    int                 `json:"page"`
	Size    int                 `json:"size"`
	Domains []mse.GatewayDomain `json:"domains"`
}

type BatchUpdateResult struct {
	AccountID        string                `json:"account_id"`
	GatewayID        string                `json:"gateway_id"`
	RequestedVersion string                `json:"requested_version"`
	DryRun           bool                  `json:"dry_run"`
	ScannedCount     int                   `json:"scanned_count"`
	WouldUpdateCount int                   `json:"would_update_count"`
	UpdatedCount     int                   `json:"updated_count"`
	FailedCount      int                   `json:"failed_count"`
	Errors           []string              `json:"errors"`
	Changes          []DomainUpdatePreview `json:"changes,omitempty"`
}

type BatchUpdateSummary struct {
	AccountID        string                `json:"account_id"`
	GatewayID        string                `json:"gateway_id,omitempty"`
	RequestedVersion string                `json:"requested_version,omitempty"`
	DryRun           bool                  `json:"dry_run"`
	ScannedCount     int                   `json:"scanned_count"`
	WouldUpdateCount int                   `json:"would_update_count"`
	UpdatedCount     int                   `json:"updated_count"`
	FailedCount      int                   `json:"failed_count"`
	Errors           []string              `json:"errors"`
	Results          []BatchUpdateResult   `json:"results,omitempty"`
	Changes          []DomainUpdatePreview `json:"changes,omitempty"`
}

type DomainConfigSnapshot struct {
	Protocol  string `json:"protocol"`
	TLSMin    string `json:"tls_min"`
	TLSMax    string `json:"tls_max"`
	MustHTTPS bool   `json:"must_https"`
	HTTP2     string `json:"http2"`
	CertID    string `json:"cert_id"`
}

type DomainUpdatePreview struct {
	Domain string               `json:"domain"`
	Before DomainConfigSnapshot `json:"before"`
	After  DomainConfigSnapshot `json:"after"`
}

type SingleUpdateResult struct {
	AccountID        string               `json:"account_id"`
	GatewayID        string               `json:"gateway_id"`
	Domain           string               `json:"domain"`
	RequestedVersion string               `json:"requested_version"`
	DryRun           bool                 `json:"dry_run"`
	Changed          bool                 `json:"changed"`
	Preview          *DomainUpdatePreview `json:"preview,omitempty"`
}

type AccountInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GatewayInfo struct {
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	ID        string `json:"id"`
	BatchTLS  string `json:"batch_tls_version"`
}

type GatewayCertificateInfo struct {
	CertIdentifier string `json:"cert_identifier"`
	CertName       string `json:"cert_name"`
	CertExpireAt   string `json:"cert_expire_at"`
	BoundDomain    string `json:"bound_domain"`
}

type AuditQuery struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	AccountID string `json:"account_id"`
	GatewayID string `json:"gateway_id"`
	Action    string `json:"action"`
	Result    string `json:"result"`
}

type AuditPage struct {
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
	Logs  []audit.Entry `json:"logs"`
}

func NewDomainService(clientManager *mse.ClientManager, cfg *config.Config, auditLogger *audit.Logger, store *store.Store) *DomainService {
	return &DomainService{clientManager: clientManager, cfg: cfg, auditLogger: auditLogger, store: store}
}

func (s *DomainService) ListDomains(ctx context.Context, accountID, gatewayID string, page, pageSize int, searchType, searchKeyword string) (*DomainPage, error) {
	client, gateway, err := s.resolveAccountGateway(accountID, gatewayID)
	if err != nil {
		return nil, err
	}
	domains, err := client.ListGatewayDomains(ctx, gateway.ID)
	if err != nil {
		return nil, err
	}
	certs, certErr := client.ListGatewayCertificates(ctx, gateway.ID)
	if certErr != nil {
		fmt.Printf("[WARN] ListGatewayCertificates failed: gateway=%s err=%v\n", gateway.ID, certErr)
	} else {
		certMap := make(map[string]mse.GatewayCertificate, len(certs))
		for _, cert := range certs {
			if cert.CertIdentifier == "" {
				continue
			}
			certMap[cert.CertIdentifier] = cert
		}
		for i := range domains {
			if cert, ok := certMap[domains[i].CertIdentifier]; ok {
				domains[i].CertName = cert.CertName
				domains[i].CertExpireAt = cert.ExpireAt
			}
		}
	}

	fmt.Printf("[DEBUG] ListDomains: searchType=%s, searchKeyword=%s, original count=%d\n", searchType, searchKeyword, len(domains))

	if searchKeyword != "" {
		var filtered []mse.GatewayDomain
		for _, d := range domains {
			match := false
			lowerKeyword := strings.ToLower(searchKeyword)
			switch searchType {
			case "domain":
				match = strings.Contains(strings.ToLower(d.Name), lowerKeyword)
				fmt.Printf("[DEBUG] domain match: name=%s, keyword=%s, match=%v\n", d.Name, lowerKeyword, match)
			case "protocol":
				match = strings.Contains(strings.ToLower(d.Protocol), lowerKeyword)
				fmt.Printf("[DEBUG] protocol match: protocol=%s, keyword=%s, match=%v\n", d.Protocol, lowerKeyword, match)
			case "tls":
				match = strings.Contains(strings.ToLower(d.TLSMin), lowerKeyword) || strings.Contains(strings.ToLower(d.TLSMax), lowerKeyword)
				fmt.Printf("[DEBUG] tls match: TLSMin=%s, TLSMax=%s, keyword=%s, match=%v\n", d.TLSMin, d.TLSMax, lowerKeyword, match)
			case "cert":
				certKeyword := strings.ToLower(d.CertName)
				if certKeyword == "" {
					certKeyword = strings.ToLower(d.CertIdentifier)
				}
				match = certKeyword != "" && strings.Contains(certKeyword, lowerKeyword)
				fmt.Printf("[DEBUG] cert match: CertName=%s, CertIdentifier=%s, keyword=%s, match=%v\n", d.CertName, d.CertIdentifier, lowerKeyword, match)
			default:
				match = strings.Contains(strings.ToLower(d.Name), lowerKeyword)
				fmt.Printf("[DEBUG] default match: name=%s, keyword=%s, match=%v\n", d.Name, lowerKeyword, match)
			}
			if match {
				filtered = append(filtered, d)
			}
		}
		domains = filtered
		fmt.Printf("[DEBUG] ListDomains: filtered count=%d\n", len(domains))
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	total := len(domains)
	start := (page - 1) * pageSize
	if start >= total {
		return &DomainPage{Total: total, Page: page, Size: pageSize, Domains: []mse.GatewayDomain{}}, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return &DomainPage{
		Total:   total,
		Page:    page,
		Size:    pageSize,
		Domains: domains[start:end],
	}, nil
}

func (s *DomainService) ListAccounts() ([]AccountInfo, error) {
	accounts, err := s.store.ListAccounts()
	if err != nil {
		return nil, err
	}
	result := make([]AccountInfo, 0, len(accounts))
	for _, acct := range accounts {
		result = append(result, AccountInfo{
			ID:   acct.ID,
			Name: acct.Name,
		})
	}
	return result, nil
}

func (s *DomainService) ListGateways(accountID string) ([]GatewayInfo, error) {
	gateways, err := s.store.ListGateways(accountID)
	if err != nil {
		return nil, err
	}
	result := make([]GatewayInfo, 0, len(gateways))
	for _, gw := range gateways {
		batchTLS := gw.BatchTLS
		if batchTLS == "" {
			batchTLS = s.cfg.Update.BatchTLS
		}
		result = append(result, GatewayInfo{
			AccountID: gw.AccountID,
			Name:      gw.Name,
			ID:        gw.ID,
			BatchTLS:  batchTLS,
		})
	}
	return result, nil
}

func (s *DomainService) ListGatewayCertificates(ctx context.Context, accountID, gatewayID string) ([]GatewayCertificateInfo, error) {
	client, gateway, err := s.resolveAccountGateway(accountID, gatewayID)
	if err != nil {
		return nil, err
	}
	certs, err := client.ListGatewayCertificates(ctx, gateway.ID)
	if err != nil {
		return nil, err
	}
	result := make([]GatewayCertificateInfo, 0, len(certs))
	for _, cert := range certs {
		result = append(result, GatewayCertificateInfo{
			CertIdentifier: cert.CertIdentifier,
			CertName:       cert.CertName,
			CertExpireAt:   cert.ExpireAt,
			BoundDomain:    cert.BoundDomain,
		})
	}
	return result, nil
}

func (s *DomainService) BatchUpdate(ctx context.Context, accountID, gatewayID, protocol, version string, dryRun bool, mustHTTPS string, http2 string) (*BatchUpdateSummary, error) {
	if strings.EqualFold(strings.TrimSpace(accountID), "all") {
		return s.batchUpdateAll(ctx, protocol, version, dryRun, mustHTTPS, http2)
	}
	result, err := s.batchUpdateSingle(ctx, accountID, gatewayID, protocol, version, dryRun, mustHTTPS, http2)
	if err != nil {
		return nil, err
	}
	return &BatchUpdateSummary{
		AccountID:        result.AccountID,
		GatewayID:        result.GatewayID,
		RequestedVersion: result.RequestedVersion,
		DryRun:           result.DryRun,
		ScannedCount:     result.ScannedCount,
		WouldUpdateCount: result.WouldUpdateCount,
		UpdatedCount:     result.UpdatedCount,
		FailedCount:      result.FailedCount,
		Errors:           result.Errors,
		Changes:          result.Changes,
	}, nil
}

func (s *DomainService) batchUpdateSingle(ctx context.Context, accountID, gatewayID, protocol, version string, dryRun bool, mustHTTPS string, http2 string) (*BatchUpdateResult, error) {
	client, gateway, err := s.resolveAccountGateway(accountID, gatewayID)
	if err != nil {
		return nil, err
	}
	resolvedProtocol, err := s.resolveProtocol(protocol)
	if err != nil {
		return nil, err
	}
	requestedVersion := ""
	tlsMin := ""
	tlsMax := ""
	if resolvedProtocol != "HTTP" {
		requestedVersion = s.resolveVersion(gateway.BatchTLS, version)
		if requestedVersion != "" {
			tlsValue, err := toMSETLS(requestedVersion)
			if err != nil {
				return nil, err
			}
			tlsMin = tlsValue
			tlsMax = tlsValue
		}
	}

	domains, err := client.ListGatewayDomains(ctx, gateway.ID)
	if err != nil {
		return nil, err
	}

	result := &BatchUpdateResult{
		AccountID:        accountID,
		GatewayID:        gateway.ID,
		RequestedVersion: requestedVersion,
		DryRun:           dryRun,
		ScannedCount:     len(domains),
		Errors:           make([]string, 0),
		Changes:          make([]DomainUpdatePreview, 0),
	}
	for _, domain := range domains {
		targetProtocol := s.resolveTargetProtocol(domain.Protocol, resolvedProtocol)
		target := s.buildTargetDomain(domain, targetProtocol, mustHTTPS, http2, "")
		nextTLSMin := domain.TLSMin
		nextTLSMax := domain.TLSMax
		if targetProtocol == "HTTPS" && tlsMin != "" {
			nextTLSMin = tlsMin
			nextTLSMax = tlsMax
		}
		preview := s.buildDomainUpdatePreview(domain, target, nextTLSMin, nextTLSMax)
		if !s.hasDomainConfigChange(preview) {
			continue
		}
		result.WouldUpdateCount++
		if dryRun {
			result.Changes = append(result.Changes, preview)
			continue
		}
		result.Changes = append(result.Changes, preview)

		if err := client.UpdateDomainTLS(ctx, gateway.ID, target, nextTLSMin, nextTLSMax); err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, err.Error())
			s.writeAudit(accountID, "batch_update", gateway.ID, domain.Name, requestedVersion, dryRun, "failed", err.Error(), &preview)
			continue
		}
		result.UpdatedCount++
		s.writeAudit(accountID, "batch_update", gateway.ID, domain.Name, requestedVersion, dryRun, "success", "", &preview)
	}
	return result, nil
}

func (s *DomainService) batchUpdateAll(ctx context.Context, protocol, version string, dryRun bool, mustHTTPS string, http2 string) (*BatchUpdateSummary, error) {
	summary := &BatchUpdateSummary{
		AccountID: "all",
		DryRun:    dryRun,
		Errors:    make([]string, 0),
		Results:   make([]BatchUpdateResult, 0),
		Changes:   make([]DomainUpdatePreview, 0),
	}
	accounts, err := s.store.ListAccounts()
	if err != nil {
		return nil, err
	}
	for _, acct := range accounts {
		gateways, err := s.store.ListGateways(acct.ID)
		if err != nil {
			summary.FailedCount++
			summary.Errors = append(summary.Errors, fmt.Sprintf("account=%s err=%v", acct.ID, err))
			continue
		}
		for _, gw := range gateways {
			result, err := s.batchUpdateSingle(ctx, acct.ID, gw.ID, protocol, version, dryRun, mustHTTPS, http2)
			if err != nil {
				summary.FailedCount++
				summary.Errors = append(summary.Errors, fmt.Sprintf("account=%s gateway=%s err=%v", acct.ID, gw.ID, err))
				continue
			}
			summary.Results = append(summary.Results, *result)
			summary.ScannedCount += result.ScannedCount
			summary.WouldUpdateCount += result.WouldUpdateCount
			summary.UpdatedCount += result.UpdatedCount
			summary.FailedCount += result.FailedCount
			summary.Errors = append(summary.Errors, result.Errors...)
			summary.Changes = append(summary.Changes, result.Changes...)
		}
	}
	if len(summary.Results) == 0 && len(summary.Errors) > 0 {
		return summary, fmt.Errorf("batch update for all accounts failed")
	}
	return summary, nil
}

func (s *DomainService) UpdateOneDomain(ctx context.Context, accountID, gatewayID, domainName, protocol, version, certIdentifier string, dryRun bool, mustHTTPS string, http2 string) (*SingleUpdateResult, error) {
	targetName := strings.TrimSpace(strings.ToLower(domainName))
	if targetName == "" {
		return nil, fmt.Errorf("domain is required")
	}

	client, gateway, err := s.resolveAccountGateway(accountID, gatewayID)
	if err != nil {
		return nil, err
	}
	resolvedProtocol, err := s.resolveProtocol(protocol)
	if err != nil {
		return nil, err
	}
	requestedVersion := ""
	tlsMin := ""
	tlsMax := ""
	if resolvedProtocol != "HTTP" {
		requestedVersion = s.resolveVersion(gateway.BatchTLS, version)
		if requestedVersion != "" {
			tlsValue, err := toMSETLS(requestedVersion)
			if err != nil {
				return nil, err
			}
			tlsMin = tlsValue
			tlsMax = tlsValue
		}
	}

	domains, err := client.ListGatewayDomains(ctx, gateway.ID)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if strings.ToLower(domain.Name) != targetName {
			continue
		}
		targetProtocol := s.resolveTargetProtocol(domain.Protocol, resolvedProtocol)
		target := s.buildTargetDomain(domain, targetProtocol, mustHTTPS, http2, certIdentifier)
		nextTLSMin := domain.TLSMin
		nextTLSMax := domain.TLSMax
		if targetProtocol == "HTTPS" && tlsMin != "" {
			nextTLSMin = tlsMin
			nextTLSMax = tlsMax
		}
		preview := s.buildDomainUpdatePreview(domain, target, nextTLSMin, nextTLSMax)
		changed := s.hasDomainConfigChange(preview)

		result := &SingleUpdateResult{
			AccountID:        accountID,
			GatewayID:        gateway.ID,
			Domain:           domain.Name,
			RequestedVersion: requestedVersion,
			DryRun:           dryRun,
			Changed:          changed,
			Preview:          &preview,
		}

		if dryRun {
			s.writeAudit(accountID, "single_update", gateway.ID, domain.Name, requestedVersion, true, "success", "dry-run", &preview)
			return result, nil
		}
		if !changed {
			s.writeAudit(accountID, "single_update", gateway.ID, domain.Name, requestedVersion, false, "success", "no-change", &preview)
			return result, nil
		}
		if err := client.UpdateDomainTLS(ctx, gateway.ID, target, nextTLSMin, nextTLSMax); err != nil {
			s.writeAudit(accountID, "single_update", gateway.ID, domain.Name, requestedVersion, false, "failed", err.Error(), &preview)
			return nil, err
		}
		s.writeAudit(accountID, "single_update", gateway.ID, domain.Name, requestedVersion, false, "success", "", &preview)
		return result, nil
	}
	return nil, fmt.Errorf("domain %s does not belong to configured gateway in account %s", domainName, accountID)
}

func (s *DomainService) QueryAuditLogs(q AuditQuery) (*AuditPage, error) {
	if s.auditLogger == nil {
		return &AuditPage{Total: 0, Page: 1, Size: 20, Logs: []audit.Entry{}}, nil
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 200 {
		q.PageSize = 200
	}
	q.AccountID = strings.TrimSpace(q.AccountID)
	q.GatewayID = strings.TrimSpace(q.GatewayID)
	q.Action = strings.TrimSpace(q.Action)
	q.Result = strings.TrimSpace(q.Result)

	logs, total, err := s.auditLogger.Query(q.AccountID, q.GatewayID, q.Action, q.Result, q.Page, q.PageSize)
	if err != nil {
		return nil, err
	}

	return &AuditPage{
		Total: int(total),
		Page:  q.Page,
		Size:  q.PageSize,
		Logs:  logs,
	}, nil
}

func SupportedTLSVersions() []string {
	return []string{"tlsv1.0", "tlsv1.2", "tlsv1.3"}
}

func toMSETLS(version string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(version))
	mseValue, ok := tlsMap[normalized]
	if !ok {
		return "", fmt.Errorf("unsupported tls version: %s", version)
	}
	return mseValue, nil
}

func (s *DomainService) buildTargetDomain(domain mse.GatewayDomain, protocol, mustHTTPS, http2, certIdentifier string) mse.GatewayDomain {
	target := domain
	target.Protocol = protocol
	if strings.TrimSpace(certIdentifier) != "" {
		target.CertIdentifier = strings.TrimSpace(certIdentifier)
	}

	if protocol != "HTTPS" {
		target.MustHTTPS = false
		target.Http2 = "close"
		return target
	}

	if mustHTTPS == "true" {
		target.MustHTTPS = true
	} else if mustHTTPS == "false" {
		target.MustHTTPS = false
	}
	if http2 != "" && http2 != "default" {
		target.Http2 = http2
	}
	return target
}

func (s *DomainService) buildDomainUpdatePreview(before mse.GatewayDomain, afterDomain mse.GatewayDomain, tlsMin, tlsMax string) DomainUpdatePreview {
	return DomainUpdatePreview{
		Domain: before.Name,
		Before: DomainConfigSnapshot{
			Protocol:  before.Protocol,
			TLSMin:    before.TLSMin,
			TLSMax:    before.TLSMax,
			MustHTTPS: before.MustHTTPS,
			HTTP2:     before.Http2,
			CertID:    before.CertIdentifier,
		},
		After: DomainConfigSnapshot{
			Protocol:  afterDomain.Protocol,
			TLSMin:    tlsMin,
			TLSMax:    tlsMax,
			MustHTTPS: afterDomain.MustHTTPS,
			HTTP2:     afterDomain.Http2,
			CertID:    afterDomain.CertIdentifier,
		},
	}
}

func (s *DomainService) hasDomainConfigChange(preview DomainUpdatePreview) bool {
	return preview.Before.Protocol != preview.After.Protocol ||
		preview.Before.TLSMin != preview.After.TLSMin ||
		preview.Before.TLSMax != preview.After.TLSMax ||
		preview.Before.MustHTTPS != preview.After.MustHTTPS ||
		preview.Before.HTTP2 != preview.After.HTTP2 ||
		preview.Before.CertID != preview.After.CertID
}

func (s *DomainService) resolveProtocol(protocol string) (string, error) {
	value := strings.TrimSpace(strings.ToUpper(protocol))
	if value == "" || value == "DEFAULT" || value == "KEEP" {
		return "DEFAULT", nil
	}
	if value == "HTTP" || value == "HTTPS" {
		return value, nil
	}
	return "", fmt.Errorf("unsupported protocol: %s", protocol)
}

func (s *DomainService) resolveTargetProtocol(domainProtocol, selectedProtocol string) string {
	if selectedProtocol != "DEFAULT" {
		return selectedProtocol
	}
	value := strings.TrimSpace(strings.ToUpper(domainProtocol))
	if value == "HTTP" || value == "HTTPS" {
		return value
	}
	return "HTTPS"
}

func (s *DomainService) resolveAccountGateway(accountID, gatewayID string) (*mse.Client, *store.Gateway, error) {
	selectedAccount := strings.TrimSpace(accountID)
	selectedGateway := strings.TrimSpace(gatewayID)

	if selectedAccount == "" {
		accounts, err := s.store.ListAccounts()
		if err != nil || len(accounts) == 0 {
			return nil, nil, fmt.Errorf("no accounts configured")
		}
		selectedAccount = accounts[0].ID
	}

	client, err := s.clientManager.GetClient(selectedAccount)
	if err != nil {
		return nil, nil, err
	}

	if selectedGateway == "" {
		gateways, err := s.store.ListGateways(selectedAccount)
		if err != nil || len(gateways) == 0 {
			return nil, nil, fmt.Errorf("no gateways configured for account %s", selectedAccount)
		}
		selectedGateway = gateways[0].ID
	}

	gateway, err := s.store.GetGatewayByID(selectedGateway)
	if err != nil {
		return nil, nil, fmt.Errorf("gateway %s is not configured under account %s", selectedGateway, selectedAccount)
	}

	return client, gateway, nil
}

func (s *DomainService) resolveVersion(gatewayBatchTLS, version string) string {
	_ = gatewayBatchTLS
	if version != "" {
		return strings.TrimSpace(strings.ToLower(version))
	}
	return ""
}

func (s *DomainService) writeAudit(accountID, action, gatewayID, domain, tlsVersion string, dryRun bool, result, message string, preview *DomainUpdatePreview) {
	if s.auditLogger == nil {
		return
	}
	entry := audit.Entry{
		AccountID:  accountID,
		Action:     action,
		GatewayID:  gatewayID,
		Domain:     domain,
		TLSVersion: tlsVersion,
		DryRun:     dryRun,
		Result:     result,
		Message:    message,
	}
	if preview != nil {
		beforeMustHTTPS := preview.Before.MustHTTPS
		afterMustHTTPS := preview.After.MustHTTPS
		entry.CertBefore = preview.Before.CertID
		entry.CertAfter = preview.After.CertID
		entry.Protocol = preview.After.Protocol
		entry.ProtocolBefore = preview.Before.Protocol
		entry.ProtocolAfter = preview.After.Protocol
		entry.TLSBefore = preview.Before.TLSMin
		entry.TLSAfter = preview.After.TLSMin
		entry.MustHTTPS = preview.After.MustHTTPS
		entry.MustHTTPSBefore = &beforeMustHTTPS
		entry.MustHTTPSAfter = &afterMustHTTPS
		entry.HTTP2 = preview.After.HTTP2
		entry.HTTP2Before = preview.Before.HTTP2
		entry.HTTP2After = preview.After.HTTP2
	}
	_ = s.auditLogger.Write(entry)
}

func (s *DomainService) RemoveAccountClient(accountID string) {
	s.clientManager.RemoveClient(accountID)
}
