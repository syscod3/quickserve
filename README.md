# quickserve

`quickserve` is a tiny Go CLI that serves a directory over HTTP and prints URLs for local and LAN access. It can also request an opt-in UPnP router mapping when you explicitly ask for public sharing.

## Security Warning

`quickserve` has no TLS and no authentication. Anyone who can reach the server can read files under the selected directory. When `-upnp` succeeds, those files may be reachable from the public internet.

Only serve directories you intend to share.

## Install From GitHub Releases

Download the archive for your platform from:

https://github.com/syscod3/quickserve/releases/latest

macOS Apple Silicon:

```bash
curl -LO https://github.com/syscod3/quickserve/releases/download/v0.1.2/quickserve_v0.1.2_darwin_arm64.tar.gz
tar -xzf quickserve_v0.1.2_darwin_arm64.tar.gz
install -m 0755 quickserve /opt/homebrew/bin/quickserve
```

macOS Intel:

```bash
curl -LO https://github.com/syscod3/quickserve/releases/download/v0.1.2/quickserve_v0.1.2_darwin_amd64.tar.gz
tar -xzf quickserve_v0.1.2_darwin_amd64.tar.gz
install -m 0755 quickserve /usr/local/bin/quickserve
```

Linux amd64:

```bash
curl -LO https://github.com/syscod3/quickserve/releases/download/v0.1.2/quickserve_v0.1.2_linux_amd64.tar.gz
tar -xzf quickserve_v0.1.2_linux_amd64.tar.gz
sudo install -m 0755 quickserve /usr/local/bin/quickserve
```

Linux arm64:

```bash
curl -LO https://github.com/syscod3/quickserve/releases/download/v0.1.2/quickserve_v0.1.2_linux_arm64.tar.gz
tar -xzf quickserve_v0.1.2_linux_arm64.tar.gz
sudo install -m 0755 quickserve /usr/local/bin/quickserve
```

Windows amd64:

Download `quickserve_v0.1.2_windows_amd64.zip`, extract `quickserve.exe`, and place it in a directory on your `PATH`.

## Install With Go

```bash
go install github.com/syscod3/quickserve@latest
```

## Basic Use

Serve the current directory on port `8000`:

```bash
quickserve
```

Serve another directory:

```bash
quickserve -dir ~/Downloads
```

Choose a port:

```bash
quickserve -port 9000
```

Let the OS choose a free port:

```bash
quickserve -port 0
```

## UPnP

UPnP is disabled by default. Enable it only when you want to ask your router to expose the server.

```bash
quickserve -dir ~/Public -upnp
```

Use a different external port:

```bash
quickserve -dir ~/Public -port 8000 -upnp -upnp-port 18080
```

Request a shorter temporary lease:

```bash
quickserve -upnp -upnp-lease 30m
```

Request a permanent mapping:

```bash
quickserve -upnp -upnp-lease 0
```

Temporary mappings are renewed while `quickserve` runs. On `Ctrl-C` or `SIGTERM`, it removes only the mapping created by that process.

## Flags

```text
-dir string
      directory to serve (default ".")
-port int
      local TCP port; use 0 to select an available port (default 8000)
-upnp
      request a public TCP port mapping using UPnP IGD
-upnp-port int
      external UPnP port; 0 uses the selected local port
-upnp-lease duration
      UPnP lease duration; 0 requests a permanent mapping (default 1h0m0s)
-version
      print version information and exit
```

## Example Output

```text
Serving: /Users/giovanni/Downloads
Local:   http://localhost:8000/
LAN:     http://192.168.1.42:8000/
Public:  http://203.0.113.10:8000/
WARNING: This HTTP server has no TLS or authentication. Serve only files you intend to share.
         It binds to all interfaces intentionally for LAN/public serving.
```

## Network Notes

`quickserve` binds to all interfaces intentionally so other devices on the LAN can connect. macOS may ask whether to allow incoming network connections; allow them if LAN access is required.

Public access may still fail when UPnP succeeds. Common causes are double NAT, carrier-grade NAT, firewall policy, ISP filtering, or a router that accepts a mapping but does not route inbound traffic correctly.

The server uses Go's standard `http.FileServer`. A selected root must be a valid directory. Directory listings use the standard Go behavior. Symlinks inside the served root follow normal filesystem behavior, so do not serve a directory containing symlinks to files you do not intend to share.

## Verify Checksums

Download `checksums.txt` and your archive, then run:

```bash
shasum -a 256 -c checksums.txt --ignore-missing
```

On Windows, use:

```powershell
Get-FileHash .\quickserve_v0.1.2_windows_amd64.zip -Algorithm SHA256
```

Compare the hash to `checksums.txt`.

## Verify Artifact Attestation

Install the GitHub CLI, then run:

```bash
gh attestation verify --owner syscod3 quickserve_v0.1.2_darwin_arm64.tar.gz
```

## Build From Source

```bash
git clone https://github.com/syscod3/quickserve.git
cd quickserve
go build ./...
go build -o quickserve .
```

## Supported Platforms

Release binaries are published for:

- macOS arm64
- macOS amd64
- Linux arm64
- Linux amd64
- Windows amd64

## License

`quickserve` is released under CC0-1.0.

It depends on `github.com/huin/goupnp` for UPnP IGD discovery and SOAP calls.
