import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatExperimentDateTime } from "../lib/format-experiment-date-time";
import type { ExperimentBriefing } from "../services/get-experiment-briefing-service";

type ExperimentBriefCardProps = {
  briefing?: ExperimentBriefing;
  hasRefreshError: boolean;
  isRefreshing: boolean;
};

export function ExperimentBriefCard({
  briefing,
  hasRefreshError,
  isRefreshing,
}: ExperimentBriefCardProps) {
  const brief = briefing?.latestBrief;
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
              <BriefField label="判断" value={brief.decision} />
              <BriefField label="仮説" value={brief.hypothesis} />
              <BriefField label="成功基準" value={brief.successCriteria} />
              <BriefField label="必要条件" value={brief.requiredConditions} />
              <BriefField label="未解決事項" value={brief.openQuestion} />
            </dl>
          )}
        </CardContent>
      </Card>
      <p className="text-sm text-muted-foreground">
        最終確認: {formatExperimentDateTime(briefing?.lastConfirmedAt)}
      </p>
    </aside>
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
