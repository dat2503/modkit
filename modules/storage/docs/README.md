# Storage Module

Object storage for modkit projects using AWS S3 (or any S3-compatible provider).

## Overview

The storage module wraps S3-compatible object storage to provide file upload, download, signed URLs, and deletion. Compatible with AWS S3, Cloudflare R2, and MinIO out of the box.

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `s3` | AWS S3 | MVP | Go, Bun |

## Setup

**AWS S3:**
1. Create an S3 bucket in your AWS account
2. Create an IAM user with `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` permissions on the bucket
3. Set env vars with the IAM user credentials

**Cloudflare R2 (recommended — zero egress costs):**
1. Create an R2 bucket in your Cloudflare account
2. Create an R2 API token with read/write permissions
3. Set `STORAGE_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com` and `STORAGE_REGION=auto`

## Configuration

See `config.schema.json` for all environment variables.

## Dependencies

- **observability** (optional) — traces upload/download calls
