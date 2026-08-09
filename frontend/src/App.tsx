import { ExperimentListPage } from "@/features/experiments/ExperimentListPage";
import { getExperimentBriefing } from "@/features/experiments/services/get-experiment-briefing-service";
import { listExperiments } from "@/features/experiments/services/list-experiments-service";
import { startExperimentBriefing } from "@/features/experiments/services/start-experiment-briefing-service";

export default function App() {
  return (
    <ExperimentListPage
      listExperiments={listExperiments}
      getExperimentBriefing={getExperimentBriefing}
      startExperimentBriefing={startExperimentBriefing}
    />
  );
}
