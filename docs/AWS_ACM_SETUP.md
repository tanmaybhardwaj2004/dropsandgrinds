# AWS ACM (Certificate Manager) Setup

## Overview
AWS Certificate Manager (ACM) provides SSL/TLS certificates for secure HTTPS connections.

## ACM Configuration Steps

### 1. Request a Public Certificate
```bash
aws acm request-certificate \
  --domain-name dropsandgrinds.com \
  --subject-alternative-names *.dropsandgrinds.com \
  --validation-method DNS
```

### 2. Validate the Certificate
After requesting, ACM will provide DNS validation records. Add these to your domain's DNS configuration.

**Example DNS Records:**
```
Name: _a8b9c0d1e2f3.dropsandgrinds.com
Type: CNAME
Value: _x2y3z4a5b6c7.certificates.amazonaws.com

Name: _x2y3z4a5b6c7.dropsandgrinds.com
Type: CNAME
Value: _a8b9c0d1e2f3.certificates.amazonaws.com
```

### 3. Wait for Validation
Monitor certificate status:
```bash
aws acm describe-certificate \
  --certificate-arn arn:aws:acm:region:account:certificate/xxx \
  --query "Certificate.Status"
```

Status will change from `PENDING_VALIDATION` to `ISSUED` once DNS records propagate.

### 4. Use Certificate with ALB
Update ALB listener to use the ACM certificate:
```bash
aws elbv2 modify-listener \
  --listener-arn arn:aws:elasticloadbalancing:region:account:listener/app/dropsandgrinds-alb/xxx/xxx \
  --certificates CertificateArn=arn:aws:acm:region:account:certificate/xxx
```

### 5. Use Certificate with CloudFront
Update CloudFront distribution:
```bash
aws cloudfront update-distribution \
  --id DISTRIBUTION_ID \
  --distribution-config '{
    "ViewerCertificate": {
      "ACMCertificateArn": "arn:aws:acm:region:account:certificate/xxx",
      "SSLSupportMethod": "sni-only",
      "MinimumProtocolVersion": "TLSv1.2_2021"
    }
  }'
```

## SSL/TLS Configuration

### Security Policy
Use modern TLS configuration:
- **Protocol**: TLS 1.2 and TLS 1.3 only
- **Cipher suites**: Modern, secure ciphers
- **Forward secrecy**: Enabled

### Certificate Renewal
ACM automatically renews certificates before expiration. Ensure DNS validation records remain in place.

## Security Best Practices

1. **Use DNS Validation**: More secure than email validation
2. **Enable Perfect Forward Secrecy**: Use ECDHE key exchange
3. **Disable Weak Protocols**: Disable SSLv3, TLS 1.0, TLS 1.1
4. **Use Strong Cipher Suites**: Prefer AES-GCM over CBC
5. **Enable HSTS**: Add HTTP Strict Transport Security header
6. **Monitor Expiration**: ACM auto-renews, but monitor for issues

## HSTS Configuration
Add HSTS header to application middleware:
```go
w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
```

## Troubleshooting

### Certificate Not Issued
- Verify DNS records are correctly configured
- Check DNS propagation (can take up to 48 hours)
- Ensure CNAME records point to correct AWS endpoints

### Certificate Expired
- ACM should auto-renew
- Check if DNS validation records were removed
- Verify certificate is in use by resources

### TLS Handshake Failures
- Check security policy compatibility
- Verify certificate chain is complete
- Check client supports configured TLS version
