## Architecture Overview

ProcGuard is a dual-mode Windows application built with Wails v2 that combines a GUI frontend with a background Native Messaging Host for browser integration.

### Dual-Mode Execution

The same executable (`ProcGuard.exe`) operates in two distinct modes based on launch arguments [1](#5-0) :

1. **GUI Mode**: User-facing application with Wails WebView interface
2. **Native Messaging Host Mode**: Background process for browser extension communication

Mode selection occurs in `main()` by checking if `os.Args[1]` starts with `chrome-extension://` [2](#5-1) .

### Core Components

```
src/
├── main.go              # Entry point, mode detection, Wails app setup
├── bindings.go          # App struct embedding API server
├── api/                 # Backend API layer (exposed to frontend)
├── internal/            # Core business logic
└── frontend/            # Svelte/TypeScript UI
```

### Initialization Flow

GUI mode initialization sequence in `startup()` [3](#5-2) :
1. Save Wails runtime context
2. Initialize SQLite database
3. Create multi-target logger
4. Instantiate API server
5. Start background daemon
6. Register native messaging host

### Key Architectural Patterns

- **Async Write Queue**: All database writes go through a buffered channel to prevent blocking [4](#5-3) 
- **Process Isolation**: GUI and Native Host run as separate processes, communicating via SQLite and heartbeat files
- **Platform Abstraction**: Windows-specific code isolated in `internal/platform/` with build tags
