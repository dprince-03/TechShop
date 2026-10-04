# AWS Deployment Patterns

| Need | Simple option | Scalable option |
|---|---|---|
| Container app | App Runner / Lightsail containers | ECS Fargate behind ALB |
| VM | EC2 + Nginx + systemd/Docker | EC2 Auto Scaling + ALB |
| Static frontend | S3 + CloudFront (OAC) | same + WAF |
| PostgreSQL/MySQL | RDS (Multi-AZ for prod) | Aurora |
| Redis | ElastiCache | ElastiCache cluster mode |
| Secrets | SSM Parameter Store | Secrets Manager (rotation) |
| Files | S3 private + presigned URLs | + CloudFront signed URLs |
| Email | SES (with SPF/DKIM/DMARC) | |
| DNS/TLS | Route 53 + ACM | |

Baseline:
- Separate accounts (or at least VPCs) for prod and non-prod; IAM Identity Center for humans; roles for workloads.
- RDS/ElastiCache in private subnets; security groups reference each other, not CIDRs.
- CloudTrail, GuardDuty, billing alarms on.
- Infrastructure in Terraform/CDK; remote state in S3 with locking.
- ECS: task role (app permissions) separate from execution role (pull image, read secrets); secrets injected via `secrets` from SSM/Secrets Manager.
