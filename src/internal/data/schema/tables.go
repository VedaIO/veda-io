package schema

const AppSchema = `
	-- app_events stores information about running processes.
	CREATE TABLE IF NOT EXISTS app_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		process_name TEXT NOT NULL,
		pid INTEGER NOT NULL,
		parent_process_name TEXT,
		exe_path TEXT,
		start_time INTEGER NOT NULL,
		end_time INTEGER,
		process_instance_key TEXT
	);

	-- Indexes to speed up queries on app_events.
	CREATE INDEX IF NOT EXISTS idx_app_events_start_time ON app_events (start_time);
	CREATE INDEX IF NOT EXISTS idx_app_events_end_time ON app_events (end_time);
	CREATE INDEX IF NOT EXISTS idx_app_events_pid ON app_events (pid);

	-- web_events stores the URLs of visited websites.
	CREATE TABLE IF NOT EXISTS web_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		domain TEXT NOT NULL,
		timestamp INTEGER NOT NULL
	);

	-- Index to speed up queries on web_events.
	CREATE INDEX IF NOT EXISTS idx_web_events_timestamp ON web_events (timestamp);
	CREATE INDEX IF NOT EXISTS idx_web_events_domain ON web_events (domain);

	-- logs stores application logs.
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL
	);

	-- Index to speed up queries on logs.
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs (timestamp);

	-- web_metadata stores cached metadata for websites (title, icon).
	CREATE TABLE IF NOT EXISTS web_metadata (
		domain TEXT PRIMARY KEY,
		title TEXT,
		icon_url TEXT,
		timestamp INTEGER NOT NULL
	);

	-- screen_time stores foreground window usage time.
	CREATE TABLE IF NOT EXISTS screen_time (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		executable_path TEXT,
		timestamp INTEGER NOT NULL,
		duration_seconds INTEGER DEFAULT 1
	);

	-- Indexes for screen_time queries.
	CREATE INDEX IF NOT EXISTS idx_screen_time_timestamp ON screen_time (timestamp);
	CREATE INDEX IF NOT EXISTS idx_screen_time_exe ON screen_time (executable_path);
`
