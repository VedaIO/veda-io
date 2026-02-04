# ProcGuard

ProcGuard is a Windows-based tool for monitoring and controlling processes and web activity on your system. It is composed of a daemon that runs in the background, an API server with a web-based GUI, and a browser extension for web monitoring.

## Features

- **Process Monitoring:** Logs all running processes and their activity.
- **Application Blocking:** Block any application from running.
- **Web Activity Monitoring:** Logs all visited websites.
- **Website Blocking:** Block any website from being accessed.
- **Web-based GUI:** A simple and intuitive web interface to view logs and manage blocklists.
- **Browser Extension:** A Chrome extension for web monitoring and blocking.

### How to build

1. Pre-exquisites: All the packages in flake.nix (best use direnv) and wails cli.

Install wails cli:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

2. Build the app:

```bash
cd src
make build
```
