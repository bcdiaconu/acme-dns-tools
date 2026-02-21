# dns-proxy-tools

dns-proxy is a Go project for managing DNS TXT records via cPanel, supporting both an HTTP API and a CLI tool for secure automation (e.g., Let's Encrypt DNS-01 challenges).

> **Note:** This project uses the `cPanel API 2` (ZoneEdit module) for all DNS operations. See cPanel documentation for details.

## Features

- HTTP API (`dns-proxy-api`): Exposes `/set_txt` endpoint for remote TXT record management
- CLI tool (`dns-proxy-cli`): Allows local DNS TXT record management via command line, ideal for certbot hooks
- Reads configuration from `/etc/acme-dns-tools/dns-proxy-api.conf` (API) or `/etc/acme-dns-tools/dns-proxy-cli.conf` (CLI)

## Configuration

Create a config file for each app:

- For the HTTP API (`dns-proxy-api`): `/etc/acme-dns-tools/dns-proxy-api.conf`

  ```ini
  DNS_RESOLVER_API_TOKEN=your_api_token_here
  ```

- For the CLI (`dns-proxy-cli`): `/etc/dns-proxy-cli.conf`
  
  ```ini
  cpanel_url=https://your-cpanel-domain:2083
  cpanel_user=cpanel_username
  cpanel_apikey=cpanel_api_token
  ```

- `DNS_RESOLVER_API_TOKEN`: The Bearer token required for API requests (only for API)
- `cpanel_url`, `cpanel_user`, `cpanel_apikey`: cPanel credentials (only for CLI)

## Build

Use the provided Makefile to build both binaries:

```sh
make
```

This will generate:

- `dns-proxy-api` (HTTP API server)
- `dns-proxy-cli` (command-line tool)

## Running as a Service (SystemD)

To run `dns-proxy-api` as a systemd service on Linux:

1. Create a systemd service file `/etc/systemd/system/dns-proxy-api.service` with the following content:

   ```ini
   [Unit]
   Description=DNS Proxy API Service
   After=network.target

   [Service]
   Type=simple
   ExecStart=/usr/local/bin/dns-proxy-api
   Restart=on-failure
   User=nobody
   Group=nogroup

   [Install]
   WantedBy=multi-user.target
   ```

   Adjust `User` and `Group` as needed for your environment.

1. Reload systemd and start the service:

   ```sh
   systemctl daemon-reload
   systemctl enable dns-proxy-api
   systemctl start dns-proxy-api
   ```

1. Check the status:

   ```sh
   systemctl status dns-proxy-api
   ```

### OpenRC (Alpine Linux)

For Alpine Linux (OpenRC), create `/etc/init.d/dns-proxy-api` with:

```sh
#!/sbin/openrc-run
command="/usr/local/bin/dns-proxy-api"
command_background="yes"
description="DNS Proxy API Service"

pidfile="/var/run/dns-proxy-api.pid"

start_pre() {
    checkpath --directory /var/run
}
```

Make it executable and enable/start the service:

```sh
chmod +x /etc/init.d/dns-proxy-api
rc-update add dns-proxy-api
dns-proxy-api start
```

## Usage

### HTTP API (for remote integration)

1. **Start the server:**

   ```sh
   ./dns-proxy-api
   ```

1. **Send a request to set a TXT record:**

   - Endpoint: `POST /set_txt`
   - Headers:
     - `Authorization: Bearer <API_KEY>`
     - `Content-Type: application/json`
   - Body:

     ```json
     {
       "domain": "example.com",
       "key": "_acme-challenge",
       "value": "your_txt_value"
     }
     ```

   Example using `curl`:

   ```sh
   curl -X POST http://localhost:5000/set_txt \
     -H "Authorization: Bearer your_api_key_here" \
     -H "Content-Type: application/json" \
     -d '{"domain":"example.com","key":"_acme-challenge","value":"txt_value_here"}'
   ```

### CLI (for local automation/certbot)

1. **Set a TXT record:**

   ```sh
   dns-proxy-cli set-txt --domain example.com --key _acme-challenge --value txt_value_here
   ```

1. **Example certbot hook:**

   ```sh
   dns-proxy-cli set-txt --domain "$CERTBOT_DOMAIN" --key "_acme-challenge.$CERTBOT_DOMAIN" --value "$CERTBOT_VALIDATION"
   ```

### CLI Commands

The `dns-proxy-cli` supports the following commands:

- **set-txt**: Add or update a DNS TXT record

  ```sh
  dns-proxy-cli set-txt --domain <domain> --key <key> --value <value>
  ```

  - `--domain`: The domain name (e.g., example.com)
  - `--key`: The TXT record key (e.g., _acme-challenge)
  - `--value`: The TXT record value

- **delete-txt**: Remove a DNS TXT record

  ```sh
  dns-proxy-cli delete-txt --domain <domain> --key <key> --value <value>
  ```

  - `--domain`: The domain name
  - `--key`: The TXT record key
  - `--value`: The TXT record value (must match the value to be deleted)

You can extend the CLI by adding new commands in the `internal/commands/` directory, each as a separate file implementing the `Command` interface.

## Notes

- Use the CLI for maximum security dacă rulezi totul local.
- Use the HTTP API only if you need remote access.
- Config files are separate for each binary, but can be identical in content.

## License

MIT License
