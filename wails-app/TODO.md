# ProcGuard Development Handoff / TODO

## 1. Project Context & Environment
*   **App**: ProcGuard (Windows Process Manager/Blocker).
*   **Stack**: Wails v2 (Go Backend + Svelte Frontend).
*   **Dev Environment**: WSL2 (Fedora 42) with Nix Flakes.
*   **Target**: Windows (x86_64).
*   **Build System**: **Manual CGO Cross-Compilation with Zig**.
    *   We **DO NOT** use `wails build` for the final binary because it fails to pass CGO environment variables correctly.
    *   We use a custom `Makefile` that runs `go build` with `CC="zig cc -target x86_64-windows-gnu"`.

## 2. Current Status (What's Done)
*   **CGO Integration**: Successfully set up `internal/native` modules compiling with Zig.
*   **Self-Protection (`internal/native/selfprotect`)**:
    *   Uses Windows DACL to deny `PROCESS_TERMINATE` rights.
    *   Prevents Task Manager from killing the app ("Access Denied").
    *   Active in `main.go` startup.
*   **Installer Logic (`internal/installer`)**:
    *   `Install()`: Copies exe to `Program Files`, adds **Windows Defender Exclusion** (via PowerShell), sets Registry Run key, creates Start Menu shortcut.
    *   `Uninstall()`: Removes Registry key, Shortcut, and Defender exclusion.
    *   `IsInstalled()`: Checks if running from Program Files.
    *   `IsAdmin()`: Checks for admin privileges.

## 3. Immediate Next Steps (The "To Do")

### A. Expose Installer API to Frontend
We need to bridge the Go `installer` package to the Svelte frontend.
1.  **Locate `App` struct**: It's likely in `main.go` or `app.go`.
2.  **Add Methods**:
    ```go
    func (a *App) InstallSystemWide() error { return installer.Install() }
    func (a *App) UninstallSystemWide() error { return installer.Uninstall() }
    func (a *App) IsSystemWide() bool { return installer.IsInstalled() }
    func (a *App) IsAdmin() bool { return installer.IsAdmin() }
    ```
3.  **Rebuild**: Run `make build` to regenerate Wails bindings (`frontend/wailsjs`).

### B. Implement Settings UI
Update `frontend/src/lib/Settings.svelte`:
1.  Add a "System Installation" card.
2.  **Logic**:
    *   On mount, call `IsSystemWide()` and `IsAdmin()`.
    *   If **Not Installed**: Show "Install System-Wide" button.
        *   If **Not Admin**: Disable button or show warning "Run as Admin required".
    *   If **Installed**: Show "Uninstall" button.
3.  **UX**: Show loading state during install (it might take a few seconds due to file copy/Defender commands).

### C. Testing
1.  Build: `make build` (in WSL2).
2.  Run `build/bin/ProcGuard.exe` on **Windows**.
3.  Test Self-Protection: Try to kill via Task Manager (should fail).
4.  Test Install: Click "Install", verify file in Program Files, Registry Key, and Defender Exclusion.

## 4. Important Gotchas
*   **Windows Defender**: The app is unsigned. The `installer` package adds a local exclusion to prevent "Virus detected" false positives. This only works *after* installation. The initial run might still trigger SmartScreen (User must click "Run Anyway").
*   **Wails Build**: Do NOT run `wails build`. Always use `make build` or `make build-debug`.
*   **Debug Console**: `make build-debug` keeps the console open and uses embedded assets (via `desktop,production` tag hack) to prevent path errors while allowing debugging.

## 5. Future Plans (See `future-plans/` dir)
*   **Screen Time**: Implement `internal/native/screentime` using `SetWinEventHook`.
*   **Anti-Cheat**: Implement hash-based process tracking.
