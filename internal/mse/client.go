package mse

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	sdkmse "github.com/aliyun/alibaba-cloud-sdk-go/services/mse"
)

// GatewayDomain 网关域名配置。
type GatewayDomain struct {
	ID             int64  `json:"id"`              //域名ID
	Name           string `json:"name"`            //域名名称
	Protocol       string `json:"protocol"`        //协议
	MustHTTPS      bool   `json:"must_https"`      //是否强制HTTPS
	Http2          string `json:"http2"`           //Http2配置
	CertIdentifier string `json:"cert_identifier"` //证书ID
	CertName       string `json:"cert_name"`       //证书名称
	CertExpireAt   string `json:"cert_expire_at"`  //证书过期时间
	TLSMin         string `json:"tls_min"`         //TLS最小版本
	TLSMax         string `json:"tls_max"`         //TLS最大版本
}

type GatewayCertificate struct {
	CertIdentifier string
	CertName       string
	ExpireAt       string
	BoundDomain    string
}

// Client 阿里云 MSE SDK 客户端。
type Client struct {
	raw            *sdkmse.Client
	acceptLanguage string
}

// NewClient 创建单账号对应的 MSE SDK 客户端。
func NewClient(regionID, accessKeyID, accessKeySecret, acceptLanguage string) (*Client, error) {
	raw, err := sdkmse.NewClientWithAccessKey(
		regionID,
		accessKeyID,
		accessKeySecret,
	)
	if err != nil {
		return nil, fmt.Errorf("create mse client failed: %w", err)
	}
	return &Client{
		raw:            raw,
		acceptLanguage: acceptLanguage,
	}, nil
}

// ListGatewayDomains 拉取指定网关下域名配置。
func (c *Client) ListGatewayDomains(_ context.Context, gatewayID string) ([]GatewayDomain, error) {
	req := sdkmse.CreateListGatewayDomainRequest()
	req.GatewayUniqueId = gatewayID
	req.AcceptLanguage = c.acceptLanguage

	resp, err := c.raw.ListGatewayDomain(req)
	if err != nil {
		return nil, fmt.Errorf("list gateway domains failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("list gateway domains failed: code=%d message=%s", resp.Code, resp.Message)
	}

	domains := make([]GatewayDomain, 0, len(resp.Data))
	for _, d := range resp.Data {
		domains = append(domains, GatewayDomain{
			ID:             d.Id,
			Name:           d.Name,
			Protocol:       d.Protocol,
			MustHTTPS:      d.MustHttps,
			Http2:          d.Http2,
			CertIdentifier: d.CertIdentifier,
			TLSMin:         d.TlsMin,
			TLSMax:         d.TlsMax,
		})
	}
	return domains, nil
}

func (c *Client) ListGatewayCertificates(_ context.Context, gatewayID string) ([]GatewayCertificate, error) {
	req := sdkmse.CreateListSSLCertRequest()
	req.GatewayUniqueId = gatewayID
	req.AcceptLanguage = c.acceptLanguage

	resp, err := c.raw.ListSSLCert(req)
	if err != nil {
		return nil, fmt.Errorf("list gateway ssl certs failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("list gateway ssl certs failed: code=%d message=%s", resp.Code, resp.Message)
	}

	certs := make([]GatewayCertificate, 0, len(resp.Data))
	for _, cert := range resp.Data {
		expireAt := cert.GmtAfter
		if expireAt == "" {
			expireAt = cert.AfterDate
		}
		boundDomain := cert.Sans
		if boundDomain == "" {
			boundDomain = cert.CommonName
		}
		if boundDomain == "" {
			boundDomain = cert.Name
		}
		certs = append(certs, GatewayCertificate{
			CertIdentifier: cert.CertIdentifier,
			CertName:       cert.CertName,
			ExpireAt:       expireAt,
			BoundDomain:    boundDomain,
		})
	}
	return certs, nil
}

// UpdateDomainTLS 更新单个网关域名 TLS 配置。
// 注意：UpdateGatewayDomain 需要携带协议、证书等完整参数。
func (c *Client) UpdateDomainTLS(_ context.Context, gatewayID string, domain GatewayDomain, tlsMin, tlsMax string) error {
	req := sdkmse.CreateUpdateGatewayDomainRequest()
	req.GatewayUniqueId = gatewayID
	req.Id = requests.NewInteger64(domain.ID)
	req.Protocol = domain.Protocol
	req.MustHttps = requests.NewBoolean(domain.MustHTTPS)
	req.Http2 = domain.Http2
	req.CertIdentifier = domain.CertIdentifier
	req.TlsMin = tlsMin
	req.TlsMax = tlsMax
	req.AcceptLanguage = c.acceptLanguage

	resp, err := c.raw.UpdateGatewayDomain(req)
	if err != nil {
		return fmt.Errorf("update gateway domain %s failed: %w", domain.Name, err)
	}
	if !resp.Success {
		return fmt.Errorf("update gateway domain %s failed: code=%d message=%s", domain.Name, resp.Code, resp.Message)
	}
	return nil
}
