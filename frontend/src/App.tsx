import { ExperimentListPage } from "@/features/experiments/ExperimentListPage";
import { listExperiments } from "@/features/experiments/services/list-experiments-service";

export default function App() {
  return <ExperimentListPage listExperiments={listExperiments} />;
}
