# ProcGuard Wails Application Architecture

## Overview

ProcGuard is a desktop application built with [Wails](https://wails.io/), combining a Go backend with a Svelte frontend. It features a unique **Dual-Mode Architecture** to handle both the GUI and the Chrome Extension Native Messaging Host using a single executable.

## Core Architecture: The "Dual-Mode" Executable

The `ProcGuard.exe` binary operates in two distinct modes depending on how it is launched. This is determined at the very beginning of `main.go`.

### 1. GUI Mode (Default)
*   **Launched by:** User (double-click, start menu).
*   **Behavior:**
    *   Starts the Wails GUI (WebView2).
    *   Enables `SingleInstanceLock` to prevent multiple GUI windows.
    *   Starts the `daemon` (process monitor) in a background goroutine.
    *   Connects to the SQLite database.
    *   **Communication:** Talks to the Frontend via Wails Bindings (`window.go`).

### 2. Native Messaging Host Mode (Background)
*   **Launched by:** Google Chrome (via the Extension).
*   **Trigger:** `os.Args` contains `chrome-extension://...`.
*   **Behavior:**
    *   **NO GUI:** Does not launch the Wails window.
    *   **Standard I/O:** Communicates with Chrome via `Stdin` (Read) and `Stdout` (Write).
    *   **Isolation:** Runs as a completely separate process from the GUI. It has its own memory space and variables.
    *   **Communication:**
        *   **With Chrome:** JSON messages over Stdio.
        *   **With GUI:** Shared Files (Heartbeat) and SQLite Database (Logs).

## Inter-Process Communication (IPC)

Since the GUI and the Native Host are separate processes, they cannot share memory variables. They communicate via:

1.  **SQLite Database (`procguard.db`):**
    *   Both processes connect to the same SQLite file in `%LocalAppData%\procguard\`.
    *   **Native Host:** Writes web logs (`log_url`) and metadata.
    *   **GUI:** Reads web logs for display.
    *   **Concurrency:** SQLite WAL (Write-Ahead Logging) mode is enabled to allow simultaneous reading and writing.

2.  **Heartbeat File (`extension_heartbeat`):**
    *   **Native Host:** Writes the current Unix timestamp to `%LocalAppData%\procguard\extension_heartbeat` every 2 seconds (and on every message received).
    *   **GUI:** Polls this file every 3 seconds. If the timestamp is recent (<10s), it considers the extension "Connected".

## Directory Structure

*   `main.go`: Entry point. Handles mode switching.
*   `frontend/`: Svelte 5 source code.
*   `internal/`: Go internal packages.
    *   `internal/web/`: Native Messaging Host implementation.
    *   `internal/daemon/`: Process monitoring logic.
    *   `internal/data/`: Database access and configuration.
    *   `internal/auth/`: Password hashing and authentication.
*   `api/`: Wails API handlers (methods exposed to Frontend).

## Debugging

*   **GUI Logs:** `%LocalAppData%\procguard\logs\procguard_debug.log`
*   **Native Host Logs:** `%LocalAppData%\procguard\logs\native_host.log` (Crucial for debugging extension issues)
