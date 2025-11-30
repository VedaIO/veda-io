# Internal Packages

This directory contains the core business logic of the ProcGuard application, separated into distinct packages to maintain clean architecture.

## Packages

### `web` (`internal/web`)
*   **Purpose:** Handles the Native Messaging Host logic for the Chrome Extension.
*   **Key File:** `native-messaging.go`.
*   **Responsibility:** Reading JSON messages from Stdin, writing to the database, and updating the heartbeat file.
*   **See:** [internal/web/README.md](./web/README.md) for detailed protocol documentation.

### `daemon` (`internal/daemon`)
*   **Purpose:** The background process monitor.
*   **Responsibility:**
    *   Monitors running processes (using `gopsutil`).
    *   Enforces the App Blocklist (kills forbidden processes).
    *   Logs application usage stats to the database.
    *   Runs independently of the GUI window state.

### `data` (`internal/data`)
*   **Purpose:** Data Access Layer (DAL).
*   **Responsibility:**
    *   **Database:** Manages the SQLite connection (`procguard.db`).
    *   **Schema:** Defines tables (`app_events`, `web_events`, `web_metadata`).
    *   **Config:** Manages `config.json` (password hash, settings).
    *   **Logging:** Provides a centralized file logger.
*   **Concurrency:** Uses a buffered channel (`StartDatabaseWriter`) to serialize writes and prevent SQLite locking issues.

### `auth` (`internal/auth`)
*   **Purpose:** Security utilities.
*   **Responsibility:** Password hashing (bcrypt) and verification.
