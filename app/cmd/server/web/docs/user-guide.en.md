# User guide - VaultCertsViewer (VCV)

## What is VCV?

VaultCertsViewer (VCV) is a lightweight web interface designed to visualize and monitor certificates managed by HashiCorp Vault (or OpenBao) PKI engines. It provides a centralized dashboard to track expiration dates, status (valid, expired, revoked), and technical details of your certificates across multiple Vault instances and PKI mounts.

## Capabilities

- **Multi-Vault support**: Connect to one or multiple Vault instances simultaneously.
- **PKI engine selector**: Filter certificates by Vault instance and PKI mount using an interactive modal with search, select/deselect per vault or globally.
- **Dashboard**: Donut chart with real-time statistics on certificate status distribution (valid, expiring, expired, revoked). Click a segment or status card to filter the table instantly.
- **Search & filter**: Search by Common Name (CN) or Subject Alternative Names (SAN). Filter by status via the dashboard cards.
- **Sorting**: Sort the certificate table by Common Name, Created date, Expiry date, Vault name, or PKI mount. Click a column header to toggle ascending/descending.
- **Pagination**: Server-side pagination with configurable page sizes (25, 50, 100, or All).
- **Detailed view**: Access full certificate metadata in a modal including issuer, subject, key algorithm, key usage, fingerprints (SHA-1, SHA-256), and PEM content.
- **PEM download**: Download certificate PEM files directly from the table or the detail modal.
- **Vault status**: A header indicator (shield icon with status dot) shows the live connection status of your Vault instances. Click it to open a detailed status modal with per-vault health information and a refresh button.
- **Expiration notifications**: A banner at the top of the page warns about certificates expiring within the configured thresholds (critical / warning).
- **Toast notifications**: Real-time toast messages for Vault connection changes, errors, and user feedback.
- **Cache & refresh**: Certificate data is cached server-side (15 min TTL). Use the refresh button (↻) in the header to invalidate the cache and fetch fresh data.
- **In-app documentation**: Access this user guide and the configuration reference directly from the UI via the documentation button (📖).
- **URL state sync**: Filters, search, sort order, pagination, and mount selection are reflected in the URL for bookmarking and sharing.
- **I18n**: Full support for English, French, Spanish, German, and Italian. Switch languages with the dropdown in the header.
- **Dark mode**: Modern UI with persistent dark/light mode toggle.
- **Admin panel**: Manage `settings.json` visually (add/remove Vault instances, configure thresholds, logging, CORS). Requires `VCV_ADMIN_PASSWORD` environment variable.
- **Prometheus metrics**: Expose certificate and connection metrics at `/metrics` for monitoring and alerting.

## Using the interface

### Dashboard

The dashboard displays a donut chart and four status cards (Valid, Expiring, Expired, Revoked). Click any card or chart segment to filter the certificate table by that status. A "Clear filter" button appears to reset the filter.

### PKI engine selector

Click the "PKI Engines" button in the filter bar to open the mount selector modal. Mounts are grouped by Vault instance. You can:

- Search mounts by name.
- Select/deselect individual mounts.
- Select/deselect all mounts for a specific Vault instance.
- Select/deselect all mounts globally.

The certificate table updates automatically as you toggle mounts.

### Certificate details

Click the "Details" button on any row to open a modal with full certificate metadata: status badges, expiry countdown, issuer, subject, SANs, serial number, key algorithm, fingerprints, key usage, and PEM content.

### Vault status

The shield icon in the header shows the overall Vault connection state (green = all connected, red = at least one disconnected). Click it to see per-vault status. You can force a health re-check from the modal.

## Configuration

VCV is configured primarily through a `settings.json` file. The admin panel lets you edit this file visually. See the configuration documentation for full details.

All application settings (Vault instances, expiration thresholds, logging, CORS, etc.) are defined in `settings.json`. Only two environment variables are still required:

- `VCV_ADMIN_PASSWORD`: Bcrypt hash to enable the admin panel (kept as an env var for security — it should not be stored in a file editable from the UI).
- `SETTINGS_PATH`: Path to a custom `settings.json` file (only needed if the file is not in a default location).

> **Note:** Environment variables (`VAULT_ADDRS`, `LOG_LEVEL`, etc.) are still supported as a legacy fallback when no `settings.json` is found, but using `settings.json` is the recommended approach.

## Limits & what it does NOT do

- **Read-only**: VCV is a visualization tool. It does **not** allow issuing, renewing, or revoking certificates.
- **Authentication**: VCV assumes you have provided valid tokens for the Vault instances it connects to.
- **Vault management**: It does not manage Vault policies or PKI configuration; it only reads existing data.

## Support

For issues or feature requests, please refer to the project repository.
