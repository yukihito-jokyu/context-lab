import { EnvironmentPreparationListPage } from "@/features/environment-preparation/EnvironmentPreparationListPage";
import { listPreparations } from "@/features/environment-preparation/services/list-preparations-service";
import { ComparisonPage } from "@/features/experiments/ComparisonPage";
import { CreateDerivedExperimentPage } from "@/features/experiments/CreateDerivedExperimentPage";
import { DerivationSourcePage } from "@/features/experiments/DerivationSourcePage";
import { EvaluationDetailPage } from "@/features/experiments/EvaluationDetailPage";
import { ExperimentListPage } from "@/features/experiments/ExperimentListPage";
import { ExperimentPreparationPage } from "@/features/experiments/ExperimentPreparationPage";
import { ExperimentWorkspacePage } from "@/features/experiments/ExperimentWorkspacePage";
import { InsightWorkspacePage } from "@/features/experiments/InsightWorkspacePage";
import { RunEvaluationPage } from "@/features/experiments/RunEvaluationPage";
import { createDerivedExperiment } from "@/features/experiments/services/create-derived-experiment-service";
import { createExperimentFromBrief } from "@/features/experiments/services/create-experiment-from-brief-service";
import { finalizeExperimentConclusion } from "@/features/experiments/services/finalize-experiment-conclusion-service";
import { fixExperimentConditions } from "@/features/experiments/services/fix-experiment-conditions-service";
import { getDerivationBriefing } from "@/features/experiments/services/get-derivation-briefing-service";
import { getDerivationSource } from "@/features/experiments/services/get-derivation-source-service";
import { getEvaluationDetail } from "@/features/experiments/services/get-evaluation-detail-service";
import { getExperimentBriefing } from "@/features/experiments/services/get-experiment-briefing-service";
import { getExperimentComparison } from "@/features/experiments/services/get-experiment-comparison-service";
import { getExperimentPreparation } from "@/features/experiments/services/get-experiment-preparation-service";
import { getExperimentWorkspace } from "@/features/experiments/services/get-experiment-workspace-service";
import { getInsightWorkspace } from "@/features/experiments/services/get-insight-workspace-service";
import { getRunDetail } from "@/features/experiments/services/get-run-detail-service";
import { listExperiments } from "@/features/experiments/services/list-experiments-service";
import { retryEndedRun } from "@/features/experiments/services/retry-ended-run-service";
import { saveExperimentPreparationDraft } from "@/features/experiments/services/save-experiment-preparation-draft-service";
import { sendDerivationBriefMessage } from "@/features/experiments/services/send-derivation-brief-message-service";
import { sendExperimentBriefMessage } from "@/features/experiments/services/send-experiment-brief-message-service";
import { startDerivationBriefing } from "@/features/experiments/services/start-derivation-briefing-service";
import { startExperimentBriefing } from "@/features/experiments/services/start-experiment-briefing-service";
import { startExperiment } from "@/features/experiments/services/start-experiment-service";
import { startRunEvaluation } from "@/features/experiments/services/start-run-evaluation-service";
import { stopDerivationBriefing } from "@/features/experiments/services/stop-derivation-briefing-service";
import { stopExperimentBriefing } from "@/features/experiments/services/stop-experiment-briefing-service";

function decodeExperimentID(value: string) {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

export default function App() {
  if (window.location.pathname === "/preparations") {
    return (
      <EnvironmentPreparationListPage listPreparations={listPreparations} />
    );
  }

  const preparationMatch = window.location.pathname.match(
    /^\/experiments\/([^/]+)\/preparation$/,
  );
  const workspaceMatch = window.location.pathname.match(
    /^\/experiments\/([^/]+)\/workspace$/,
  );
  const runDetailMatch = window.location.pathname.match(
    /^\/experiments\/([^/]+)\/runs\/([^/]+)$/,
  );
  const comparisonMatch = window.location.pathname.match(
    /^\/experiments\/([^/]+)\/comparison$/,
  );
  const derivationSourceMatch = window.location.pathname.match(
    /^\/experiments\/([^/]+)\/derivations$/,
  );
  const insightWorkspaceMatch = window.location.pathname.match(
    /^\/experiments\/([^/]+)\/insights$/,
  );
  const createDerivedExperimentMatch = window.location.pathname.match(
    /^\/experiments\/([^/]+)\/derivations\/create$/,
  );
  const evaluationMatch = window.location.pathname.match(
    /^\/evaluations\/([^/]+)$/,
  );

  if (evaluationMatch) {
    return (
      <EvaluationDetailPage
        evaluationId={decodeExperimentID(evaluationMatch[1])}
        getEvaluationDetail={getEvaluationDetail}
        operationId={
          new URLSearchParams(window.location.search).get("operationId") ??
          undefined
        }
      />
    );
  }

  if (runDetailMatch) {
    return (
      <RunEvaluationPage
        experimentId={decodeExperimentID(runDetailMatch[1])}
        getRunDetail={getRunDetail}
        runId={decodeExperimentID(runDetailMatch[2])}
        startRunEvaluation={startRunEvaluation}
        title="run詳細"
      />
    );
  }

  if (comparisonMatch) {
    return (
      <ComparisonPage
        experimentId={decodeExperimentID(comparisonMatch[1])}
        finalizeExperimentConclusion={finalizeExperimentConclusion}
        getExperimentComparison={getExperimentComparison}
      />
    );
  }

  if (derivationSourceMatch) {
    return (
      <DerivationSourcePage
        briefingServices={{
          startDerivationBriefing,
          sendDerivationBriefMessage,
          getDerivationBriefing,
          stopDerivationBriefing,
        }}
        experimentId={decodeExperimentID(derivationSourceMatch[1])}
        getDerivationSource={getDerivationSource}
      />
    );
  }

  if (insightWorkspaceMatch) {
    return (
      <InsightWorkspacePage
        getInsightWorkspace={getInsightWorkspace}
        initialExperimentId={decodeExperimentID(insightWorkspaceMatch[1])}
      />
    );
  }

  if (createDerivedExperimentMatch) {
    return (
      <CreateDerivedExperimentPage
        createDerivedExperiment={createDerivedExperiment}
        sourceExperimentId={decodeExperimentID(createDerivedExperimentMatch[1])}
      />
    );
  }

  if (workspaceMatch) {
    return (
      <ExperimentWorkspacePage
        experimentId={decodeExperimentID(workspaceMatch[1])}
        getExperimentWorkspace={getExperimentWorkspace}
        startExperiment={startExperiment}
        startRunEvaluation={startRunEvaluation}
        retryEndedRun={retryEndedRun}
      />
    );
  }

  if (preparationMatch) {
    return (
      <ExperimentPreparationPage
        experimentId={decodeExperimentID(preparationMatch[1])}
        getExperimentPreparation={getExperimentPreparation}
        onBackToExperimentList={() => window.location.assign("/")}
        onConditionsFixed={(experimentId, operationId) =>
          window.location.assign(
            `/experiments/${encodeURIComponent(experimentId)}/workspace?operationId=${encodeURIComponent(operationId)}`,
          )
        }
        fixExperimentConditions={fixExperimentConditions}
        saveExperimentPreparationDraft={saveExperimentPreparationDraft}
      />
    );
  }

  return (
    <ExperimentListPage
      listExperiments={listExperiments}
      getExperimentBriefing={getExperimentBriefing}
      createExperimentFromBrief={createExperimentFromBrief}
      sendExperimentBriefMessage={sendExperimentBriefMessage}
      startExperimentBriefing={startExperimentBriefing}
      stopExperimentBriefing={stopExperimentBriefing}
    />
  );
}
