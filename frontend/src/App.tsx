import { EnvironmentPreparationListPage } from "@/features/environment-preparation/EnvironmentPreparationListPage";
import { listPreparations } from "@/features/environment-preparation/services/list-preparations-service";
import { EvaluationOperationPage } from "@/features/experiments/EvaluationOperationPage";
import { ExperimentListPage } from "@/features/experiments/ExperimentListPage";
import { ExperimentPreparationPage } from "@/features/experiments/ExperimentPreparationPage";
import { ExperimentWorkspacePage } from "@/features/experiments/ExperimentWorkspacePage";
import { RunEvaluationPage } from "@/features/experiments/RunEvaluationPage";
import { createExperimentFromBrief } from "@/features/experiments/services/create-experiment-from-brief-service";
import { fixExperimentConditions } from "@/features/experiments/services/fix-experiment-conditions-service";
import { getExperimentBriefing } from "@/features/experiments/services/get-experiment-briefing-service";
import { getExperimentPreparation } from "@/features/experiments/services/get-experiment-preparation-service";
import { getExperimentWorkspace } from "@/features/experiments/services/get-experiment-workspace-service";
import { listExperiments } from "@/features/experiments/services/list-experiments-service";
import { saveExperimentPreparationDraft } from "@/features/experiments/services/save-experiment-preparation-draft-service";
import { sendExperimentBriefMessage } from "@/features/experiments/services/send-experiment-brief-message-service";
import { startExperimentBriefing } from "@/features/experiments/services/start-experiment-briefing-service";
import { startExperiment } from "@/features/experiments/services/start-experiment-service";
import { startRunEvaluation } from "@/features/experiments/services/start-run-evaluation-service";
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
  const evaluationMatch = window.location.pathname.match(
    /^\/evaluations\/([^/]+)$/,
  );

  if (evaluationMatch) {
    return (
      <EvaluationOperationPage
        evaluationId={decodeExperimentID(evaluationMatch[1])}
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
        runId={decodeExperimentID(runDetailMatch[2])}
        startRunEvaluation={startRunEvaluation}
        title="run詳細"
      />
    );
  }

  if (comparisonMatch) {
    const runId = new URLSearchParams(window.location.search).get("runId");
    return (
      <RunEvaluationPage
        experimentId={decodeExperimentID(comparisonMatch[1])}
        runId={runId ?? ""}
        startRunEvaluation={startRunEvaluation}
        title="実験比較"
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
