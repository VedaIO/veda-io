package api

import (
	"src/internal/data/query"
	"strings"
	"time"
)

// --- Types ---

type AppLeaderboardItem struct {
	Rank        int    `json:"rank"`
	Name        string `json:"name"`        // Display name (commercial name if available)
	ProcessName string `json:"processName"` // Actual process name for blocking
	Icon        string `json:"icon"`
	Count       int    `json:"count"`
}

type WebLeaderboardItem struct {
	Rank   int    `json:"rank"`
	Domain string `json:"domain"`
	Title  string `json:"title"`
	Icon   string `json:"icon"`
	Count  int    `json:"count"`
}

type ScreenTimeItem struct {
	Name            string `json:"name"`
	ExecutablePath  string `json:"executablePath"`
	Icon            string `json:"icon"`
	DurationSeconds int    `json:"durationSeconds"`
}

// --- App Usage ---

func (s *Server) GetAppLeaderboard(since, until string) ([]AppLeaderboardItem, error) {
	sinceTime, _ := query.ParseTime(since)
	untilTime, _ := query.ParseTime(until)

	records, err := query.GetAppUsageRanking(s.db, sinceTime, untilTime)
	if err != nil {
		return nil, err
	}

	leaderboard := make([]AppLeaderboardItem, 0, len(records))
	for i, r := range records {
		details := s.icons.GetAppDetails(r.ExecutablePath)
		displayName := details.CommercialName
		if displayName == "" {
			displayName = r.ProcessName
		}

		leaderboard = append(leaderboard, AppLeaderboardItem{
			Rank:        i + 1,
			Name:        displayName,
			ProcessName: r.ProcessName,
			Icon:        details.IconBase64,
			Count:       r.Count,
		})
	}
	return leaderboard, nil
}

func (s *Server) GetScreenTime() ([]ScreenTimeItem, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	records, err := query.GetScreenTimeTotals(s.db, todayStart)
	if err != nil {
		return nil, err
	}

	items := make([]ScreenTimeItem, 0, len(records))
	for _, r := range records {
		details := s.icons.GetAppDetails(r.ExecutablePath)
		name := details.CommercialName
		if name == "" {
			name = extractFileName(r.ExecutablePath)
		}

		items = append(items, ScreenTimeItem{
			Name:            name,
			ExecutablePath:  r.ExecutablePath,
			Icon:            details.IconBase64,
			DurationSeconds: r.DurationSeconds,
		})
	}
	return items, nil
}

func (s *Server) GetTotalScreenTime() (int, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	return query.GetTotalDayScreenTime(s.db, todayStart)
}

// --- Web Usage ---

func (s *Server) GetWebLeaderboard(since, until string) ([]WebLeaderboardItem, error) {
	sinceTime, _ := query.ParseTime(since)
	untilTime, _ := query.ParseTime(until)

	records, err := query.GetWebUsageRanking(s.db, sinceTime, untilTime)
	if err != nil {
		return nil, err
	}

	leaderboard := make([]WebLeaderboardItem, 0, len(records))
	for i, r := range records {
		item := WebLeaderboardItem{
			Rank:   i + 1,
			Domain: r.Domain,
			Count:  r.Count,
		}
		if meta, err := query.GetWebMetadata(s.db, r.Domain); err == nil && meta != nil {
			item.Title = meta.Title
			item.Icon = meta.IconURL
		}
		leaderboard = append(leaderboard, item)
	}
	return leaderboard, nil
}

// --- Logs & Search ---

func (s *Server) Search(queryStr, since, until string) ([][]string, error) {
	return query.SearchAppEvents(s.db, strings.ToLower(queryStr), since, until)
}

func (s *Server) GetWebLogs(queryStr, since, until string) ([][]string, error) {
	return query.GetWebLogs(s.db, queryStr, since, until)
}

// --- Utils ---

func extractFileName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '\\' || path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
