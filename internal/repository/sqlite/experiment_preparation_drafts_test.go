package sqlite

import (
	"context"
	"database/sql"
	"reflect"
	"sync"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// SQLite実験準備下書きの保存とrequest ID再利用。
func TestStoreSaveExperimentPreparationDraft(t *testing.T) {
	store := newExperimentPreparationDraftStore(t)
	first := testExperimentPreparationDraft("request-1", "experiment-1", "保存した目的")

	saved, err := store.SaveExperimentPreparationDraft(context.Background(), first)
	if err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	if saved.SavedAt.IsZero() {
		t.Fatal("SavedAt = zero, want saved time")
	}
	assertExperimentPreparationDraftPersisted(t, store, saved)

	resent := testExperimentPreparationDraft("request-1", "experiment-1", "再送時に無視する目的")
	got, err := store.SaveExperimentPreparationDraft(context.Background(), resent)
	if err != nil {
		t.Fatalf("second SaveExperimentPreparationDraft() error = %v", err)
	}
	if !reflect.DeepEqual(got, saved) {
		t.Errorf("second draft = %+v, want original snapshot %+v", got, saved)
	}
	assertExperimentPreparationDraftPersisted(t, store, saved)
}

// SQLite実験準備下書きの異なる実験へのrequest ID再利用拒否。
func TestStoreSaveExperimentPreparationDraftRejectsRequestIDForOtherExperiment(t *testing.T) {
	store := newExperimentPreparationDraftStore(t)
	if _, err := store.SaveExperimentPreparationDraft(context.Background(), testExperimentPreparationDraft("request-1", "experiment-1", "保存した目的")); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	seedExperimentPreparationDraftExperiment(t, store, "experiment-2")

	_, err := store.SaveExperimentPreparationDraft(context.Background(), testExperimentPreparationDraft("request-1", "experiment-2", "別実験の目的"))
	if !apperr.IsCode(err, apperr.CodeDraftRequestInvalid) {
		t.Errorf("SaveExperimentPreparationDraft() error = %v, want code %q", err, apperr.CodeDraftRequestInvalid)
	}
}

// SQLite実験準備下書き保存途中失敗時の全変更取消。
func TestStoreSaveExperimentPreparationDraftRollsBackOnIntermediateFailure(t *testing.T) {
	store := newExperimentPreparationDraftStore(t)
	before := readExperimentPreparationDraftState(t, store, "experiment-1", "request-1")
	if _, err := store.db.Exec("CREATE TRIGGER fail_draft_prompt BEFORE INSERT ON experiment_preparation_prompts WHEN NEW.content = '保存失敗' BEGIN SELECT RAISE(ABORT, 'prompt failure'); END"); err != nil {
		t.Fatalf("create failure trigger error = %v", err)
	}

	failed := testExperimentPreparationDraft("request-1", "experiment-1", "更新後の目的")
	failed.Prompts[1].Content = "保存失敗"
	if _, err := store.SaveExperimentPreparationDraft(context.Background(), failed); err == nil {
		t.Fatal("SaveExperimentPreparationDraft() error = nil, want insert failure")
	}

	after := readExperimentPreparationDraftState(t, store, "experiment-1", "request-1")
	if !reflect.DeepEqual(after, before) {
		t.Errorf("rollback state = %+v, want original %+v", after, before)
	}
}

// SQLite実験準備下書き同一request IDの並行再送。
func TestStoreSaveExperimentPreparationDraftConcurrently(t *testing.T) {
	dataDirectory := t.TempDir()
	store, err := Open(dataDirectory)
	if err != nil {
		t.Fatalf("Open() primary store error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() primary store error = %v", err)
		}
	})
	seedExperimentPreparationDraftExperiment(t, store, "experiment-1")
	peer, err := Open(dataDirectory)
	if err != nil {
		t.Fatalf("Open() peer store error = %v", err)
	}
	t.Cleanup(func() {
		if err := peer.Close(); err != nil {
			t.Errorf("Close() peer store error = %v", err)
		}
	})
	// 別々のdatabase/sql poolを同じDBファイルへ接続し、直列化されない実接続競合を作る。
	draft := testExperimentPreparationDraft("request-1", "experiment-1", "保存した目的")

	type result struct {
		draft domain.ExperimentPreparationDraft
		err   error
	}
	ready := make(chan struct{})
	entered := make(chan struct{}, 2)
	results := make(chan result, 2)
	var waitGroup sync.WaitGroup
	for _, savingStore := range []*Store{
		store,
		peer,
	} {
		waitGroup.Add(1)
		go func(savingStore *Store) {
			defer waitGroup.Done()
			entered <- struct{}{}
			<-ready
			saved, err := savingStore.SaveExperimentPreparationDraft(context.Background(), draft)
			results <- result{
				draft: saved,
				err:   err,
			}
		}(savingStore)
	}
	<-entered
	<-entered
	close(ready)
	waitGroup.Wait()
	close(results)

	var snapshots []domain.ExperimentPreparationDraft
	for result := range results {
		if result.err != nil {
			t.Fatalf("SaveExperimentPreparationDraft() error = %v", result.err)
		}
		snapshots = append(snapshots, result.draft)
	}
	if len(snapshots) != 2 {
		t.Fatalf("concurrent snapshots length = %d, want 2", len(snapshots))
	}
	if !reflect.DeepEqual(snapshots[0], snapshots[1]) {
		t.Errorf("concurrent snapshots = %+v, want identical snapshots", snapshots)
	}
	assertExperimentPreparationDraftPersisted(t, store, snapshots[0])
}

// newExperimentPreparationDraftStore は下書き保存用SQLite storeを生成。
func newExperimentPreparationDraftStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	seedExperimentPreparationDraftExperiment(t, store, "experiment-1")

	return store
}

// seedExperimentPreparationDraftExperiment は下書き保存対象の準備中実験を登録。
func seedExperimentPreparationDraftExperiment(t *testing.T, store *Store, experimentID string) {
	t.Helper()
	const createdAt = "2026-08-10T00:00:00Z"
	if _, err := store.db.Exec("INSERT INTO experiments (id, purpose, state, updated_at) VALUES (?, ?, ?, ?)", experimentID, "初期目的", "preparing", createdAt); err != nil {
		t.Fatalf("seed experiment error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO experiment_preparations (experiment_id, briefing_session_id, briefing_version_id, hypothesis, environment_conditions, initial_input, evaluation_criteria, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", experimentID, "session-"+experimentID, "version-"+experimentID, "初期仮説", "初期環境", "初期入力", "初期評価", createdAt, createdAt); err != nil {
		t.Fatalf("seed experiment preparation error = %v", err)
	}
	for index, content := range []string{
		"初期prompt A",
		"初期prompt B",
	} {
		if _, err := store.db.Exec("INSERT INTO experiment_preparation_prompts (experiment_id, sequence_no, content) VALUES (?, ?, ?)", experimentID, index+1, content); err != nil {
			t.Fatalf("seed experiment preparation prompt error = %v", err)
		}
	}
}

// testExperimentPreparationDraft は保存対象の編集下書きを生成。
func testExperimentPreparationDraft(requestID, experimentID, purpose string) domain.ExperimentPreparationDraft {
	hypothesis := "更新後の仮説"

	return domain.ExperimentPreparationDraft{
		RequestID:             requestID,
		ExperimentID:          experimentID,
		Purpose:               purpose,
		Hypothesis:            &hypothesis,
		EnvironmentConditions: "更新後の環境",
		InitialInput:          "更新後の入力",
		EvaluationAxes:        "更新後の評価",
		Prompts: []domain.ExperimentPreparationPrompt{
			{
				SequenceNo: 1,
				Content:    "更新後prompt A",
			},
			{
				SequenceNo: 2,
				Content:    "更新後prompt B",
			},
		},
	}
}

// assertExperimentPreparationDraftPersisted は下書きsnapshotが全関連テーブルへ保存済みか検証。
func assertExperimentPreparationDraftPersisted(t *testing.T, store *Store, want domain.ExperimentPreparationDraft) {
	t.Helper()
	state := readExperimentPreparationDraftState(t, store, want.ExperimentID, want.RequestID)
	if state.purpose != want.Purpose || state.environmentConditions != want.EnvironmentConditions || state.initialInput != want.InitialInput || state.evaluationAxes != want.EvaluationAxes {
		t.Errorf("saved values = (%q, %q, %q, %q), want draft values", state.purpose, state.environmentConditions, state.initialInput, state.evaluationAxes)
	}
	if !reflect.DeepEqual(state.hypothesis, nullableHypothesis(want.Hypothesis)) {
		t.Errorf("saved hypothesis = %+v, want %+v", state.hypothesis, nullableHypothesis(want.Hypothesis))
	}
	wantPrompts := make([]string, 0, len(want.Prompts))
	for _, prompt := range want.Prompts {
		wantPrompts = append(wantPrompts, prompt.Content)
	}
	if !reflect.DeepEqual(state.promptContents, wantPrompts) {
		t.Errorf("saved prompt order = %q, want %q", state.promptContents, wantPrompts)
	}
	if state.operationCount != 1 {
		t.Errorf("draft operation count = %d, want 1", state.operationCount)
	}
	saved, found, err := findDraftOperation(context.Background(), store.db, want.RequestID)
	if err != nil {
		t.Fatalf("find saved draft operation error = %v", err)
	}
	if !found {
		t.Fatal("saved draft operation = not found, want snapshot")
	}
	if !reflect.DeepEqual(saved, want) {
		t.Errorf("saved draft operation = %+v, want %+v", saved, want)
	}
}

// nullableHypothesis は任意仮説のSQLite scan値を返す。
func nullableHypothesis(hypothesis *string) sql.NullString {
	if hypothesis == nil {
		return sql.NullString{}
	}

	return sql.NullString{
		String: *hypothesis,
		Valid:  true,
	}
}

// experimentPreparationDraftState はtransaction rollback検証用の永続化snapshot。
type experimentPreparationDraftState struct {
	purpose               string
	experimentUpdatedAt   string
	hypothesis            sql.NullString
	environmentConditions string
	initialInput          string
	evaluationAxes        string
	preparationUpdatedAt  string
	promptContents        []string
	operationCount        int
}

// readExperimentPreparationDraftState は実験準備と指定request IDの永続化状態を読む。
func readExperimentPreparationDraftState(t *testing.T, store *Store, experimentID, requestID string) experimentPreparationDraftState {
	t.Helper()
	var state experimentPreparationDraftState
	if err := store.db.QueryRow("SELECT e.purpose, e.updated_at, p.hypothesis, p.environment_conditions, p.initial_input, p.evaluation_criteria, p.updated_at FROM experiments e JOIN experiment_preparations p ON p.experiment_id = e.id WHERE e.id = ?", experimentID).Scan(&state.purpose, &state.experimentUpdatedAt, &state.hypothesis, &state.environmentConditions, &state.initialInput, &state.evaluationAxes, &state.preparationUpdatedAt); err != nil {
		t.Fatalf("query experiment preparation state error = %v", err)
	}
	rows, err := store.db.Query("SELECT content FROM experiment_preparation_prompts WHERE experiment_id = ? ORDER BY sequence_no", experimentID)
	if err != nil {
		t.Fatalf("query experiment preparation prompts error = %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			t.Fatalf("scan experiment preparation prompt error = %v", err)
		}
		state.promptContents = append(state.promptContents, content)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate experiment preparation prompts error = %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM experiment_preparation_draft_operations WHERE request_id = ?", requestID).Scan(&state.operationCount); err != nil {
		t.Fatalf("query draft operation count error = %v", err)
	}

	return state
}
