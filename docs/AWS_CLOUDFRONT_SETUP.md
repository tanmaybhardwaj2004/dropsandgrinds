# AWS CloudFront CDN Setup

## Overview
CloudFront distributes frontend static assets globally, reducing latency and improving load times.

## CloudFront Configuration Steps

### 1. Create S3 Bucket for Static Assets
```bash
aws s3api create-bucket \
  --bucket dropsandgrinds-static \
  --region us-east-1 \
  --create-bucket-configuration LocationConstraint=us-east-1
```

### 2. Configure S3 Bucket Policy
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowCloudFrontAccess",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudfront.amazonaws.com"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::dropsandgrinds-static/*",
      "Condition": {
        "StringEquals": {
          "AWS:SourceArn": "arn:aws:cloudfront::account-id:distribution/DISTRIBUTION_ID"
        }
      }
    }
  ]
}
```

### 3. Upload Frontend Assets to S3
```bash
aws s3 sync frontend/ s3://dropsandgrinds-static/ --delete
```

### 4. Create CloudFront Distribution
```bash
aws cloudfront create-distribution \
  --distribution-config '{
    "CallerReference": "dropsandgrinds-$(date +%s)",
    "Comment": "DropsAndGrinds frontend CDN",
    "DefaultCacheBehavior": {
      "TargetOriginId": "S3-dropsandgrinds-static",
      "ViewerProtocolPolicy": "redirect-to-https",
      "AllowedMethods": ["GET", "HEAD", "OPTIONS"],
      "CachedMethods": ["GET", "HEAD"],
      "ForwardedValues": {
        "QueryString": false,
        "Cookies": {
          "Forward": "none"
        }
      },
      "MinTTL": 0,
      "DefaultTTL": 86400,
      "MaxTTL": 31536000,
      "Compress": true
    },
    "Origins": {
      "Items": [
        {
          "Id": "S3-dropsandgrinds-static",
          "DomainName": "dropsandgrinds-static.s3.amazonaws.com",
          "S3OriginConfig": {
            "OriginAccessIdentity": ""
          }
        }
      ]
    },
    "DefaultRootObject": "index.html",
    "Enabled": true,
    "ViewerCertificate": {
      "ACMCertificateArn": "arn:aws:acm:region:account:certificate/xxx",
      "SSLSupportMethod": "sni-only",
      "MinimumProtocolVersion": "TLSv1.2_2021"
    },
    "PriceClass": "PriceClass_All"
  }'
```

### 5. Create Origin Access Control (OAC)
```bash
aws cloudfront create-origin-access-control \
  --origin-access-control-config '{
    "Name": "dropsandgrinds-oac",
    "OriginAccessControlOriginType": "s3",
    "SigningBehavior": "always",
    "SigningProtocol": "sigv4"
  }'
```

### 6. Update Distribution with OAC
Update the origin configuration to use the OAC instead of public access.

### 7. Update S3 Bucket Policy for OAC
Replace the CloudFront service principal with the OAC ARN in the bucket policy.

## Cache Behavior Settings

### Static Assets (CSS, JS, Images)
- TTL: 1 year (31536000 seconds)
- Compress: true
- Query strings: forward none
- Cookies: forward none

### HTML Files
- TTL: 0 seconds (no caching)
- Compress: true
- Query strings: forward all
- Cookies: forward all

### API Requests
- Bypass CloudFront (route directly to backend via ALB)

## Custom Error Pages
Configure CloudFront to return index.html for 403 and 404 errors (SPA routing):
```bash
aws cloudfront create-distribution \
  --custom-error-responses '{
    "Quantity": 2,
    "Items": [
      {
        "ErrorCode": 403,
        "ResponsePagePath": "/index.html",
        "ResponseCode": 200,
        "ErrorCachingMinTTL": 0
      },
      {
        "ErrorCode": 404,
        "ResponsePagePath": "/index.html",
        "ResponseCode": 200,
        "ErrorCachingMinTTL": 0
      }
    ]
  }'
```

## Invalidation
After deploying new frontend assets, invalidate the CloudFront cache:
```bash
aws cloudfront create-invalidation \
  --distribution-id DISTRIBUTION_ID \
  --paths "/*"
```

## Environment Variables
Update frontend to use CloudFront URL:
```
FRONTEND_URL=https://cdn.dropsandgrinds.com
API_URL=https://api.dropsandgrinds.com
```

## Security Best Practices
1. Use OAC instead of public S3 access
2. Enable HTTPS only
3. Use ACM certificates for TLS
4. Set appropriate cache TTLs
5. Enable WAF for additional protection (optional)

## Monitoring
Enable CloudFront metrics:
- Requests
- Bytes transferred
- 4xx/5xx error rates
- Latency
- Cache hit ratio

## Troubleshooting

### 403 Errors
- Check OAC configuration
- Verify S3 bucket policy
- Ensure distribution is deployed

### Cache Not Updating
- Create invalidation after deployment
- Check cache behavior settings
- Verify file paths match

### Slow Performance
- Check origin latency
- Review cache hit ratio
- Consider edge locations optimization
