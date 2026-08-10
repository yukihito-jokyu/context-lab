import { EnvironmentPreparationListPage } from "@/features/environment-preparation/EnvironmentPreparationListPage";
import { listPreparations } from "@/features/environment-preparation/services/list-preparations-service";
import { ExperimentListPage } from "@/features/experiments/ExperimentListPage";
import { ExperimentPreparationPage } from "@/features/experiments/ExperimentPreparationPage";
import { ExperimentWorkspacePage } from "@/features/experiments/ExperimentWorkspacePage";
import { createExperimentFromBrief } from "@/features/experiments/services/create-experiment-from-brief-service";
import { fixExperimentConditions } from "@/features/experiments/services/fix-experiment-conditions-service";
import { getExperimentBriefing } from "@/features/experiments/services/get-experiment-briefing-service";
import { getExperimentPreparation } from "@/features/experiments/services/get-experiment-preparation-service";
import { listExperiments } from "@/features/experiments/services/list-experiments-service";
import { saveExperimentPreparationDraft } from "@/features/experiments/services/save-experiment-preparation-draft-service";
import { sendExperimentBriefMessage } from "@/features/experiments/services/send-experiment-brief-message-service";
import { startExperimentBriefing } from "@/features/experiments/services/start-experiment-briefing-service";
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

  if (workspaceMatch) {
    return (
      <ExperimentWorkspacePage
        experimentId={decodeExperimentID(workspaceMatch[1])}
        operationId={
          new URLSearchParams(window.location.search).get("operationId") ??
          undefined
        }
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
