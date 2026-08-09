import { ExperimentListPage } from "@/features/experiments/ExperimentListPage";
import { createExperimentFromBrief } from "@/features/experiments/services/create-experiment-from-brief-service";
import { getExperimentBriefing } from "@/features/experiments/services/get-experiment-briefing-service";
import { listExperiments } from "@/features/experiments/services/list-experiments-service";
import { sendExperimentBriefMessage } from "@/features/experiments/services/send-experiment-brief-message-service";
import { startExperimentBriefing } from "@/features/experiments/services/start-experiment-briefing-service";

export default function App() {
  return (
    <ExperimentListPage
      listExperiments={listExperiments}
      getExperimentBriefing={getExperimentBriefing}
      createExperimentFromBrief={createExperimentFromBrief}
      sendExperimentBriefMessage={sendExperimentBriefMessage}
      startExperimentBriefing={startExperimentBriefing}
    />
  );
}
