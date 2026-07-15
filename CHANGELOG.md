# Changelog

## v0.1.1

- Fix UPnP discovery so routers exposing only WANIPConnection v1 or WANPPPConnection v1 can still be found after a WANIPConnection v2 search times out.

## v0.1.0

- Serve a selected directory over HTTP.
- Print localhost and LAN URLs, and a public address when one can be safely identified.
- Add opt-in UPnP IGD TCP port mapping with lease renewal and cleanup.
- Publish cross-platform release archives with checksums.
