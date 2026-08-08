import { ExperimentListPage } from "@/features/experiments/ExperimentListPage";
import { listExperiments } from "@/features/experiments/services/list-experiments-service";
import { startExperimentBriefing } from "@/features/experiments/services/start-experiment-briefing-service";

export default function App() {
  return (
    <ExperimentListPage
      listExperiments={listExperiments}
      startExperimentBriefing={startExperimentBriefing}
    />
  );
}
