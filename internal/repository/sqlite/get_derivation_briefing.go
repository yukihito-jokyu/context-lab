package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// GetDerivationBriefing は保存済み派生実験ブリーフを会話順で読み出す。
func (s *Store) GetDerivationBriefing(ctx context.Context, briefingSessionID string) (domain.DerivationBriefing, bool, error) {
	briefing, found, err := s.findDerivationBriefingSession(ctx, briefingSessionID)
	if err != nil || !found {
		return domain.DerivationBriefing{}, found, err
	}

	messages, latestMessageAt, err := s.listDerivationBriefingMessages(ctx, briefingSessionID)
	if err != nil {
		return domain.DerivationBriefing{}, false, err
	}
	latestSuggestion, latestSuggestionAt, err := s.findLatestDerivationBriefingSuggestion(ctx, briefingSessionID)
	if err != nil {
		return domain.DerivationBriefing{}, false, err
	}
	briefing.Messages = messages
	briefing.LatestSuggestion = latestSuggestion
	briefing.LastConfirmedAt = latestTime(briefing.LastConfirmedAt, latestMessageAt, latestSuggestionAt)

	return briefing, true, nil
}

// findDerivationBriefingSession は派生実験ブリーフセッションを取得する。
func (s *Store) findDerivationBriefingSession(ctx context.Context, briefingSessionID string) (domain.DerivationBriefing, bool, error) {
	var briefing domain.DerivationBriefing
	var createdAt string
	var updatedAt string
	err := s.db.QueryRowContext(ctx, "SELECT state, created_at, updated_at FROM preparation_sessions WHERE id=? AND kind=?", briefingSessionID, "derivation_brief").Scan(&briefing.State, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return domain.DerivationBriefing{}, false, nil
	}
	if err != nil {
		return domain.DerivationBriefing{}, false, fmt.Errorf("find derivation briefing session: %w", err)
	}
	if updatedAt == "" {
		updatedAt = createdAt
	}
	confirmedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.DerivationBriefing{}, false, fmt.Errorf("parse derivation briefing session update time: %w", err)
	}
	briefing.LastConfirmedAt = confirmedAt.UTC()

	return briefing, true, nil
}

// listDerivationBriefingMessages は利用者表示用会話を時系列順に取得する。
func (s *Store) listDerivationBriefingMessages(ctx context.Context, briefingSessionID string) (messages []domain.DerivationBriefingMessage, latestAt time.Time, err error) {
	rows, err := s.db.QueryContext(ctx, "SELECT role, content, sequence_no, created_at FROM derivation_briefing_messages WHERE preparation_session_id=? ORDER BY sequence_no ASC", briefingSessionID)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("query derivation briefing messages: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close derivation briefing message rows: %w", closeErr)
		}
	}()

	messages = make([]domain.DerivationBriefingMessage, 0)
	for rows.Next() {
		var message domain.DerivationBriefingMessage
		var createdAt string
		if err := rows.Scan(&message.Role, &message.Content, &message.SequenceNo, &createdAt); err != nil {
			return nil, time.Time{}, fmt.Errorf("scan derivation briefing message: %w", err)
		}
		message.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("parse derivation briefing message creation time: %w", err)
		}
		message.CreatedAt = message.CreatedAt.UTC()
		latestAt = latestTime(latestAt, message.CreatedAt)
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, time.Time{}, fmt.Errorf("iterate derivation briefing messages: %w", err)
	}

	return messages, latestAt, nil
}

// findLatestDerivationBriefingSuggestion は最新版の保存済み派生提案を取得する。
func (s *Store) findLatestDerivationBriefingSuggestion(ctx context.Context, briefingSessionID string) (*domain.DerivationBriefingSuggestion, time.Time, error) {
	suggestion := &domain.DerivationBriefingSuggestion{}
	var hypothesis sql.NullString
	var openQuestion sql.NullString
	var candidatePrompts string
	var createdAt string
	err := s.db.QueryRowContext(ctx, "SELECT id, version_no, purpose, decision, hypothesis, candidate_prompts, evaluation_criteria, environment_conditions, initial_input, success_criteria, required_conditions, open_question, created_at FROM derivation_briefing_suggestions WHERE preparation_session_id=? ORDER BY version_no DESC LIMIT 1", briefingSessionID).Scan(&suggestion.ID, &suggestion.VersionNo, &suggestion.Purpose, &suggestion.Decision, &hypothesis, &candidatePrompts, &suggestion.EvaluationCriteria, &suggestion.EnvironmentConditions, &suggestion.InitialInput, &suggestion.SuccessCriteria, &suggestion.RequiredConditions, &openQuestion, &createdAt)
	if err == sql.ErrNoRows {
		return nil, time.Time{}, nil
	}
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("find latest derivation briefing suggestion: %w", err)
	}
	if hypothesis.Valid {
		suggestion.Hypothesis = &hypothesis.String
	}
	if openQuestion.Valid {
		suggestion.OpenQuestion = &openQuestion.String
	}
	if err := json.Unmarshal([]byte(candidatePrompts), &suggestion.CandidatePrompts); err != nil {
		return nil, time.Time{}, fmt.Errorf("unmarshal derivation briefing candidate prompts: %w", err)
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("parse derivation briefing suggestion creation time: %w", err)
	}
	suggestion.CreatedAt = parsedCreatedAt.UTC()

	return suggestion, suggestion.CreatedAt, nil
}
