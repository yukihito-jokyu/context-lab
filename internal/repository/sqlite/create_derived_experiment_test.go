package sqlite

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

func TestStoreCreateDerivedExperiment(t *testing.T) {
	store, sourceID := finalizedDerivedExperimentSource(t)
	purpose := "派生した目的"
	prompts := []domain.ExperimentPreparationPrompt{
		{
			SequenceNo: 1,
			Content:    "派生prompt",
		},
	}
	changes := domain.DerivedExperimentChanges{
		Purpose: &purpose,
		Prompts: &prompts,
	}
	created, wasCreated, err := store.CreateDerivedExperiment(context.Background(), "derive-request", sourceID, changes, "比較するため")
	if err != nil || !wasCreated {
		t.Fatalf("CreateDerivedExperiment() = (%+v, %v, %v), want created", created, wasCreated, err)
	}
	var gotPurpose, state, parent, prompt string
	if err := store.db.QueryRow("SELECT purpose, state, derived_from_experiment_id FROM experiments WHERE id=?", created.ExperimentID).Scan(&gotPurpose, &state, &parent); err != nil {
		t.Fatalf("query created experiment: %v", err)
	}
	if gotPurpose != purpose || state != "preparing" || parent != sourceID {
		t.Errorf("created experiment = (%q, %q, %q), want (%q, preparing, %q)", gotPurpose, state, parent, purpose, sourceID)
	}
	if err := store.db.QueryRow("SELECT content FROM experiment_preparation_prompts WHERE experiment_id=?", created.ExperimentID).Scan(&prompt); err != nil {
		t.Fatalf("query created prompt: %v", err)
	}
	if prompt != "派生prompt" {
		t.Errorf("created prompt = %q, want derived prompt", prompt)
	}
	var sourcePurpose string
	if err := store.db.QueryRow("SELECT purpose FROM experiments WHERE id=?", sourceID).Scan(&sourcePurpose); err != nil {
		t.Fatalf("query source experiment: %v", err)
	}
	if sourcePurpose != "目的" {
		t.Errorf("source purpose = %q, want unchanged fixed source", sourcePurpose)
	}
	replayed, wasCreated, err := store.CreateDerivedExperiment(context.Background(), "derive-request", sourceID, changes, "比較するため")
	if err != nil || wasCreated || replayed.ExperimentID != created.ExperimentID {
		t.Errorf("replay = (%+v, %v, %v), want persisted result", replayed, wasCreated, err)
	}
	otherPurpose := "別の目的"
	if _, _, err := store.CreateDerivedExperiment(context.Background(), "derive-request", sourceID, domain.DerivedExperimentChanges{Purpose: &otherPurpose}, "比較するため"); !apperr.IsCode(err, apperr.CodeDerivedExperimentRequestConflict) {
		t.Errorf("different payload error = %v, want request conflict", err)
	}
	if _, _, err := store.CreateDerivedExperiment(context.Background(), "no-diff", sourceID, domain.DerivedExperimentChanges{Purpose: stringPointer("目的")}, "比較するため"); !apperr.IsCode(err, apperr.CodeDerivedExperimentInvalid) {
		t.Errorf("same changes error = %v, want invalid", err)
	}
	if _, _, err := store.CreateDerivedExperiment(context.Background(), "missing", "missing", changes, "比較するため"); !apperr.IsCode(err, apperr.CodeDerivedExperimentSourceNotFound) {
		t.Errorf("missing source error = %v, want not found", err)
	}
}

func TestStoreCreateDerivedExperimentRequiresFinalizedSource(t *testing.T) {
	store := newExperimentPreparationDraftStore(t)
	seedExperimentPreparationDraftExperiment(t, store, "source")
	draft := testExperimentPreparationDraft("draft", "source", "目的")
	if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	fixed, err := store.FixExperimentConditions(context.Background(), fixedConditionsFromDraft(draft))
	if err != nil {
		t.Fatalf("FixExperimentConditions() error = %v", err)
	}
	purpose := "派生"
	if _, _, err = store.CreateDerivedExperiment(context.Background(), "derive", fixed.ExperimentID, domain.DerivedExperimentChanges{Purpose: &purpose}, "比較するため"); !apperr.IsCode(err, apperr.CodeDerivedExperimentSourceNotEligible) {
		t.Errorf("non-finalized source error = %v, want not eligible", err)
	}
}

func TestDerivedExperimentHelpers(t *testing.T) {
	value := "変更"
	current := "元"
	source := derivedExperimentSource{
		Purpose:               "元",
		Hypothesis:            &current,
		EnvironmentConditions: "環境",
		InitialInput:          "入力",
		EvaluationAxes:        "評価",
		Prompts: []domain.ExperimentPreparationPrompt{
			{
				SequenceNo: 1,
				Content:    "prompt",
			},
		},
	}
	for _, changes := range []domain.DerivedExperimentChanges{
		{Purpose: &value},
		{Hypothesis: &value},
		{EnvironmentConditions: &value},
		{InitialInput: &value},
		{EvaluationAxes: &value},
		{
			Prompts: &[]domain.ExperimentPreparationPrompt{
				{
					SequenceNo: 1,
					Content:    "変更",
				},
			},
		},
	} {
		if !hasMaterialDerivedExperimentChange(source, changes) {
			t.Errorf("hasMaterialDerivedExperimentChange(%+v) = false, want true", changes)
		}
	}
	if hasMaterialDerivedExperimentChange(source, domain.DerivedExperimentChanges{Purpose: stringPointer("元")}) {
		t.Error("hasMaterialDerivedExperimentChange() = true, want false for same source")
	}
	payload, err := canonicalDerivedExperimentPayload(domain.DerivedExperimentChanges{Purpose: &value}, "理由")
	if err != nil || payload == "" {
		t.Fatalf("canonicalDerivedExperimentPayload() = (%q, %v), want payload", payload, err)
	}
	existing := derivedExperimentOperation{
		DerivedExperiment: domain.DerivedExperiment{
			ExperimentID:       "derived",
			SourceExperimentID: "source",
		},
		CanonicalPayload: payload,
	}
	if _, _, err := derivedExperimentReplay(existing, "other", payload); !apperr.IsCode(err, apperr.CodeDerivedExperimentRequestConflict) {
		t.Errorf("different source replay error = %v, want conflict", err)
	}
	if !isDerivedExperimentRequestConflict(assertedUniqueError{}) {
		t.Error("isDerivedExperimentRequestConflict() = false, want true")
	}
}

type assertedUniqueError struct{}

func (assertedUniqueError) Error() string {
	return "UNIQUE constraint failed: experiment_derived_operations.request_id"
}

func TestStoreCreateDerivedExperimentConvergesConcurrentRequestsAcrossStores(t *testing.T) {
	for range 20 {
		dataDirectory := t.TempDir()
		first, err := Open(dataDirectory)
		if err != nil {
			t.Fatalf("first Open() error = %v", err)
		}
		t.Cleanup(func() { _ = first.Close() })
		second, err := Open(dataDirectory)
		if err != nil {
			t.Fatalf("second Open() error = %v", err)
		}
		t.Cleanup(func() { _ = second.Close() })
		sourceID := seedFinalizedDerivedExperimentSource(t, first)
		purpose := "並行派生"
		ready, entered := make(chan struct{}), make(chan struct{}, 2)
		results := make(chan domain.DerivedExperiment, 2)
		errors := make(chan error, 2)
		var group sync.WaitGroup
		for _, store := range []*Store{
			first,
			second,
		} {
			group.Add(1)
			go func(store *Store) {
				defer group.Done()
				entered <- struct{}{}
				<-ready
				result, _, err := store.CreateDerivedExperiment(context.Background(), "concurrent-derive", sourceID, domain.DerivedExperimentChanges{Purpose: &purpose}, "比較するため")
				if err != nil {
					errors <- err
					return
				}
				results <- result
			}(store)
		}
		<-entered
		<-entered
		close(ready)
		group.Wait()
		close(errors)
		close(results)
		for err := range errors {
			t.Fatalf("concurrent CreateDerivedExperiment() error = %v", err)
		}
		var values []domain.DerivedExperiment
		for result := range results {
			values = append(values, result)
		}
		if len(values) != 2 || !reflect.DeepEqual(values[0], values[1]) {
			t.Errorf("concurrent values = %+v, want identical snapshots", values)
		}
		var operations int
		if err := first.db.QueryRow("SELECT COUNT(*) FROM experiment_derived_operations WHERE request_id=?", "concurrent-derive").Scan(&operations); err != nil {
			t.Fatalf("count operations: %v", err)
		}
		if operations != 1 {
			t.Errorf("operations = %d, want 1", operations)
		}
	}
}

func finalizedDerivedExperimentSource(t *testing.T) (*Store, string) {
	t.Helper()
	store, fixed := finalizedConclusionStore(t)
	if _, _, err := store.FinalizeExperimentConclusion(context.Background(), "finalize-source", fixed.ExperimentID, "結論"); err != nil {
		t.Fatalf("FinalizeExperimentConclusion() error = %v", err)
	}
	return store, fixed.ExperimentID
}
func seedFinalizedDerivedExperimentSource(t *testing.T, store *Store) string {
	t.Helper()
	sourceID := "experiment-1"
	seedExperimentPreparationDraftExperiment(t, store, sourceID)
	draft := testExperimentPreparationDraft("draft", sourceID, "目的")
	if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	fixed, err := store.FixExperimentConditions(context.Background(), fixedConditionsFromDraft(draft))
	if err != nil {
		t.Fatalf("FixExperimentConditions() error = %v", err)
	}
	start, _, err := store.BeginExperiment(context.Background(), "start", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if err = store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "summary"); err != nil {
		t.Fatalf("CompleteExperimentRun() error = %v", err)
	}
	evaluation, _, err := store.BeginRunEvaluation(context.Background(), "evaluation", start.Runs[0].ID)
	if err != nil {
		t.Fatalf("BeginRunEvaluation() error = %v", err)
	}
	if err = store.CompleteRunEvaluation(context.Background(), evaluation.EvaluationID, "summary"); err != nil {
		t.Fatalf("CompleteRunEvaluation() error = %v", err)
	}
	if _, _, err = store.FinalizeExperimentConclusion(context.Background(), "finalize", sourceID, "結論"); err != nil {
		t.Fatalf("FinalizeExperimentConclusion() error = %v", err)
	}
	return sourceID
}
func stringPointer(value string) *string { return &value }
