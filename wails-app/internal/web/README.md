# Native Messaging Host (`internal/web`)

## Overview

This package implements the **Native Messaging Host** protocol to communicate with the ProcGuard Chrome Extension. It runs as a separate process launched by Chrome.

## Protocol

Communication uses **Standard I/O (stdio)**.
*   **Input (Stdin):** 4-byte length prefix (uint32, little-endian) + JSON Message.
*   **Output (Stdout):** 4-byte length prefix (uint32, little-endian) + JSON Message.

### Message Format

**Request (Extension -> Host):**
```json
{
  "type": "message_type",
  "payload": ...
}
```

**Response (Host -> Extension):**
```json
{
  "type": "response_type",
  "payload": ...
}
```

## Supported Messages

### `log_url`
*   **Payload:**
    ```json
    {
      "url": "https://example.com",
      "title": "Example Domain",
      "visitTime": 1732968000
    }
    ```
*   **Action:** Logs the visit to the `web_events` table in SQLite.

### `log_web_metadata`
*   **Payload:**
    ```json
    {
      "domain": "example.com",
      "title": "Example Domain",
      "iconUrl": "https://example.com/favicon.ico"
    }
    ```
*   **Action:** Updates the `web_metadata` table.

### `get_web_blocklist`
*   **Payload:** `null`
*   **Response:** List of blocked domains `["facebook.com", "tiktok.com"]`.

### `ping`
*   **Payload:** `null`
*   **Response:** `{"type": "pong"}`.
*   **Side Effect:** Updates the `extension_heartbeat` file.

## The Heartbeat Mechanism

To allow the GUI to detect if the extension is active (since they are separate processes):

1.  **Continuous Pulse:** The `Run()` loop starts a background ticker that runs every **2 seconds**.
2.  **File Update:** It writes the current Unix timestamp to `%LocalAppData%\procguard\extension_heartbeat`.
3.  **GUI Check:** The GUI reads this file. If the timestamp is <10 seconds old, the extension is "Connected".

## Debugging

Since this process has no UI, it logs everything to:
`%LocalAppData%\procguard\logs\native_host.log`

**Panic Recovery:**
The `Run()` function includes a `defer recover()` block to catch and log any panics to the log file, preventing silent crashes.
