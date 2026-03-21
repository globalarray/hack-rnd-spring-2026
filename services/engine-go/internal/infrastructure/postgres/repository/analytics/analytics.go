package analytics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"sourcecraft.dev/benzo/testengine/internal/domain"
	repositorydto "sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/analytics/dto"
	servicedto "sourcecraft.dev/benzo/testengine/internal/service/analytics/dto"
)

type analyticsRepository struct {
	log *slog.Logger
	db  *sqlx.DB
}

func NewAnalyticsRepository(log *slog.Logger, db *sqlx.DB) *analyticsRepository {
	return &analyticsRepository{log: log, db: db}
}

func (r *analyticsRepository) GetSessionData(ctx context.Context, sessionID string) (*servicedto.GetSessionDataOutput, error) {
	const op = "analyticsRepository.GetSessionData"

	var sessionRecord repositorydto.SessionDataRecord
	if err := r.db.GetContext(ctx, &sessionRecord, queryGetSessionData, sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: get session: %w", op, err)
	}

	var responseRecords []repositorydto.RawClientResponseRecord
	if err := r.db.SelectContext(ctx, &responseRecords, queryGetSessionResponses, sessionID); err != nil {
		return nil, fmt.Errorf("%s: get responses: %w", op, err)
	}

	return mapSessionData(&sessionRecord, responseRecords), nil
}
