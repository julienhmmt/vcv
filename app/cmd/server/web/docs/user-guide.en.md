# User Documentation - VaultCertsViewer (VCV)

## What is VCV?

VaultCertsViewer (VCV) is a lightweight web interface designed to visualize and monitor certificates managed by HashiCorp Vault PKI engines. It provides a centralized dashboard to track expiration dates, status (valid, expired, revoked), and technical details of your certificates across multiple Vault instances and PKI mounts.

## Capabilities

- **Multi-Vault Support**: Connect to one or multiple Vault instances.
- **PKI Engine Discovery**: Automatically discovers PKI mounts you have access to.
- **Dashboard**: Real-time statistics on certificate status distribution and expiration timeline.
- **Search & Filter**: Search by Common Name (CN) or Subject Alternative Names (SAN). Filter by Vault, PKI mount, status, or expiration threshold.
- **Detailed View**: Access full certificate metadata including issuer, fingerprints, and PEM content.
- **Export**: Download certificate PEM files directly from the UI.
- **I18n**: Full support for English, French, Spanish, German, and Italian.
- **Dark Mode**: Modern UI with dark/light mode toggle.

## Configuration

VCV is configured primarily through environment variables or a `settings.json` file.

### Main Environment Variables

- `VAULT_ADDRS`: Comma-separated list of Vault addresses.
- `VCV_EXPIRE_WARNING`: Threshold in days for warning notifications (default: 30).
- `VCV_EXPIRE_CRITICAL`: Threshold in days for critical notifications (default: 7).
- `LOG_LEVEL`: Logging verbosity (info, debug, error).

## Limits & What it does NOT do

- **Read-Only**: VCV is currently a visualization tool. It does **not** allow issuing, renewing, or revoking certificates.
- **Authentication**: VCV assumes you have provided valid tokens or configured authentication for the Vault instances it connects to.
- **Vault Management**: It does not manage Vault policies or PKI configuration; it only reads existing data.

## Support

For issues or feature requests, please refer to the project repository.
