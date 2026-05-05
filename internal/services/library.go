package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/pkg/steam"
)

// LibraryService handles Steam library operations
type LibraryService struct {
	libraryRepo *repositories.LibraryRepository
	catalogRepo *repositories.CatalogRepository
	steamClient *steam.Client
	logger      *slog.Logger
}

// NewLibraryService creates a new library service
func NewLibraryService(
	libraryRepo *repositories.LibraryRepository,
	catalogRepo *repositories.CatalogRepository,
	steamClient *steam.Client,
	logger *slog.Logger,
) *LibraryService {
	return &LibraryService{
		libraryRepo: libraryRepo,
		catalogRepo: catalogRepo,
		steamClient: steamClient,
		logger:      logger,
	}
}

// ImportResult represents the result of a library import
type ImportResult struct {
	TotalGames    int    `json:"total_games"`
	ImportedCount int    `json:"imported_count"`
	Message       string `json:"message"`
}

// ImportLibrary imports Steam library for a user
func (s *LibraryService) ImportLibrary(ctx context.Context, userID int64, steamID string) (*ImportResult, error) {
	// Validate SteamID
	if !steam.ValidateSteamID(steamID) {
		return nil, fmt.Errorf("invalid Steam ID format")
	}

	// Fetch owned games from Steam API
	games, err := s.steamClient.GetOwnedGames(ctx, steamID)
	if err != nil {
		s.logger.Error("failed to fetch Steam library", "user_id", userID, "steam_id", steamID, "error", err)
		return nil, fmt.Errorf("failed to fetch Steam library: %w", err)
	}

	// Extract app IDs
	appIDs := make([]int64, len(games))
	for i, game := range games {
		appIDs[i] = game.AppID
	}

	// Import to database
	if err := s.libraryRepo.ImportOwnedGames(ctx, userID, appIDs); err != nil {
		s.logger.Error("failed to import library to database", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to import library: %w", err)
	}

	// Try to link Steam app IDs to game IDs (best effort)
	s.linkGamesToLibrary(ctx, userID, games)

	result := &ImportResult{
		TotalGames:    len(games),
		ImportedCount: len(appIDs),
		Message:       fmt.Sprintf("Successfully imported %d games from Steam library", len(appIDs)),
	}

	s.logger.Info("Steam library imported", "user_id", userID, "game_count", len(appIDs))
	return result, nil
}

// linkGamesToLibrary attempts to link Steam app IDs to game IDs in the database
// This is a best-effort operation - it matches by title similarity
func (s *LibraryService) linkGamesToLibrary(ctx context.Context, userID int64, games []steam.OwnedGame) {
	for _, game := range games {
		// Try to find matching game by title
		// In production, this would use fuzzy matching or a mapping table
		// For MVP, we'll do a simple exact match
		gameID, err := s.catalogRepo.FindGameByTitle(ctx, game.Name)
		if err != nil {
			// Log but continue - not all Steam games will be in our database
			continue
		}

		if gameID != 0 {
			// Link the Steam app ID to the game ID
			if err := s.libraryRepo.LinkSteamAppToGame(ctx, userID, game.AppID, gameID); err != nil {
				s.logger.Warn("failed to link Steam app to game", "steam_app_id", game.AppID, "game_id", gameID, "error", err)
			}
		}
	}
}

// GetLibrary returns the list of owned game IDs for a user
func (s *LibraryService) GetLibrary(ctx context.Context, userID int64) ([]int64, error) {
	return s.libraryRepo.GetOwnedGameIDs(ctx, userID)
}

// GetLibraryCount returns the count of owned games
func (s *LibraryService) GetLibraryCount(ctx context.Context, userID int64) (int, error) {
	return s.libraryRepo.GetLibraryCount(ctx, userID)
}

// IsGameOwned checks if a user owns a specific game
func (s *LibraryService) IsGameOwned(ctx context.Context, userID int64, gameID int64) (bool, error) {
	return s.libraryRepo.IsGameOwned(ctx, userID, gameID)
}

// FindMissingDLCs finds DLCs for owned base games that are not owned
func (s *LibraryService) FindMissingDLCs(ctx context.Context, userID int64) ([]int64, error) {
	ownedTitles, err := s.libraryRepo.GetOwnedGameTitles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(ownedTitles) == 0 {
		return []int64{}, nil
	}
	ownedSet := map[int64]struct{}{}
	for id := range ownedTitles {
		ownedSet[id] = struct{}{}
	}
	var missing []int64
	seen := map[int64]struct{}{}
	for _, title := range ownedTitles {
		for _, suffix := range []string{"DLC", "Season Pass", "Expansion"} {
			matches, err := s.catalogRepo.FindGamePricesByTitle(ctx, title+" "+suffix, 20)
			if err != nil {
				continue
			}
			for _, game := range matches {
				if _, owned := ownedSet[game.ID]; owned {
					continue
				}
				if _, ok := seen[game.ID]; ok {
					continue
				}
				seen[game.ID] = struct{}{}
				missing = append(missing, game.ID)
			}
		}
	}
	return missing, nil
}
