# go-shorty

URL shortener made in go.

## Scope

The end goal is to be able to run a local version of `go-shorty` and have a
local database of URLs that are associated with shortened URLs. The final design
should address the following goals:

- Users can benefit from shortened urls in the browser to save time. Example:
  `u/hn` could route to HackerNews
- Over 80% of code is tested
- Everything runs locally(docker compose or kind cluster)
- Prefer libraries over frameworks
- CI: Linting, formatting, testing, vulnerability scanning, conventional commits
- CD: Build Go binaries, build docker containers and publish to GitHub
  Packages/Releases

Here's an idea of what the target architecture would look like, this is not
fleshed out so it could change based on what's possible:

![go-shorty architecture](./.github/assets/go-shorty-arch.png)

## Manual DNS Configuration

### Windows
To enable go-shorty redirects, open Powershell with Administrator permissions:

```powershell
$iface = Get-NetAdapter | Where-Object Status -eq 'Up' | Select-Object -First 1 -Expand Name
Set-DnsClientServerAddress -InterfaceAlias $iface -ServerAddresses 127.0.0.1
ipconfig /flushdns
```

Then to restore default settings, using Powershell with Administrator:

```powershell
Set-DnsClientServerAddress -InterfaceAlias $iface -ResetServerAddresses
ipconfig /flushdns
```

### MacOS

To view your current DNS configuration, run `scutil --dns`. Then, you can create
the `/etc/resolver` folder and update a file `short` with:

```text
nameserver 127.0.0.1
port 53
```

and then the following should work:

```bash
> dig @127.0.0.1 -p 53 go.u

; <<>> DiG 9.10.6 <<>> @127.0.0.1 -p 53 go.u
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 38584
;; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;go.u.                          IN      A

;; ANSWER SECTION:
go.u.                   0       IN      A       127.0.0.1

;; Query time: 0 msec
;; SERVER: 127.0.0.1#53(127.0.0.1)
;; WHEN: Wed Jul 30 15:47:55 CDT 2025
;; MSG SIZE  rcvd: 42
```

To enable go-shorty redirects, run the following with root:

```bash
networksetup -setdnsservers "Wi‑Fi" 127.0.0.1
sudo killall -HUP mDNSResponder; sudo dscacheutil -flushcache
```

This will:

- Set the DNS server for your Wi-Fi connection to your own computer
  (localhost) meaning all DNS queries from your Mac will go to whatever DNS
  server is running on your machine (such as your go-shorty project).
- Restarts the macOS DNS resolver (mDNSResponder) and flushes the DNS cache.
  Ensures that DNS changes take effect immediately and old DNS records are
  cleared.

Then to restore default settings:

```bash
networksetup -setdnsservers "Wi‑Fi" empty
```
