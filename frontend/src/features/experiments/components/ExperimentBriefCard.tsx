import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatExperimentDateTime } from "../lib/format-experiment-date-time";
import type { ExperimentBriefing } from "../services/get-experiment-briefing-service";

type ExperimentBriefCardProps = {
  briefing?: ExperimentBriefing;
  hasRefreshError: boolean;
  isRefreshing: boolean;
  isCreating: boolean;
  onCreate: () => void;
  createError?: { code: string; message: string };
};

export function ExperimentBriefCard({
  briefing,
  hasRefreshError,
  isRefreshing,
  isCreating,
  onCreate,
  createError,
}: ExperimentBriefCardProps) {
  const brief = briefing?.latestBrief;
  const isComplete = Boolean(
    brief?.purpose?.trim() &&
      brief.candidatePrompts?.filter((prompt) => prompt.trim()).length >= 2 &&
      brief.evaluationAxes?.trim() &&
      brief.environmentConditions?.trim(),
  );
  return (
    <aside aria-labelledby="experiment-brief-title" className="space-y-3">
      <Card>
        <CardHeader>
          <CardTitle className="text-base" id="experiment-brief-title">
            実験ブリーフ（下書き）
          </CardTitle>
        </CardHeader>
        <CardContent>
          {!isRefreshing && !hasRefreshError && !brief && (
            <p className="text-sm text-muted-foreground">
              下書きはまだありません。
            </p>
          )}
          {brief && (
            <dl className="space-y-3 text-sm">
              <BriefField
                label="目的"
                value={brief.purpose ?? brief.decision}
              />
              <BriefField label="仮説" value={brief.hypothesis} />
              <BriefList label="候補prompt" values={brief.candidatePrompts} />
              <BriefField
                label="評価軸"
                value={brief.evaluationAxes ?? brief.successCriteria}
              />
              <BriefField
                label="環境条件"
                value={brief.environmentConditions ?? brief.requiredConditions}
              />
              <BriefField label="未解決事項" value={brief.openQuestion} />
            </dl>
          )}
        </CardContent>
      </Card>
      <div className="space-y-2">
        {!isComplete && brief && (
          <p
            className="text-sm text-muted-foreground"
            id="create-experiment-incomplete"
          >
            目的、候補prompt 2件、評価軸、環境条件が揃うと採用できます。
          </p>
        )}
        {isCreating && (
          <p id="create-experiment-pending" role="status">
            実験を作成しています…
          </p>
        )}
        {createError && (
          <p id="create-experiment-command-error" role="alert">
            {createError.message}
          </p>
        )}
        <button
          className="inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground disabled:pointer-events-none disabled:opacity-50"
          disabled={!isComplete || isCreating}
          id="submit-create-experiment-button"
          onClick={onCreate}
          type="button"
        >
          {isCreating ? "作成中…" : "このブリーフを採用して作成"}
        </button>
      </div>
      <p className="text-sm text-muted-foreground">
        最終確認: {formatExperimentDateTime(briefing?.lastConfirmedAt)}
      </p>
    </aside>
  );
}

function BriefList({ label, values }: { label: string; values?: string[] }) {
  if (!values?.length) return null;
  return (
    <div>
      <dt className="font-medium">{label}</dt>
      <dd className="mt-1 space-y-1 text-muted-foreground">
        {values.map((value, index) => (
          <p className="whitespace-pre-wrap break-words" key={value}>
            {index + 1}. {value}
          </p>
        ))}
      </dd>
    </div>
  );
}

function BriefField({ label, value }: { label: string; value?: string }) {
  if (!value) return null;
  return (
    <div>
      <dt className="font-medium">{label}</dt>
      <dd className="mt-1 whitespace-pre-wrap break-words text-muted-foreground">
        {value}
      </dd>
    </div>
  );
}
