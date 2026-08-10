import { AlertCircle, LoaderCircle } from "lucide-react";
import { useMemo, useRef, useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import type {
  CreatedInsight,
  CreateInsightService,
} from "../services/create-insight-service";
import type { InsightEvidenceCandidate } from "../services/get-insight-workspace-service";

function validationMessage(
  evidenceCount: number,
  statement: string,
  applicabilityConditions: string,
  verificationGaps: string,
) {
  if (evidenceCount < 2) return "異なる実験から根拠を2件以上選択してください。";
  if (!statement) return "知見文を入力してください。";
  if (!applicabilityConditions) return "適用条件を入力してください。";
  if (!verificationGaps) return "検証不足を入力してください。";
  return undefined;
}

export function CreateInsightForm({
  candidates,
  initialExperimentId,
  createInsight,
  onCreated,
}: {
  candidates: InsightEvidenceCandidate[];
  initialExperimentId?: string;
  createInsight: CreateInsightService;
  onCreated: (insight: CreatedInsight) => void;
}) {
  const [selectedConclusionIds, setSelectedConclusionIds] = useState<
    Set<string>
  >(
    () =>
      new Set(
        candidates
          .filter((candidate) => candidate.experimentId === initialExperimentId)
          .map((candidate) => candidate.conclusionId),
      ),
  );
  const [statement, setStatement] = useState("");
  const [applicabilityConditions, setApplicabilityConditions] = useState("");
  const [verificationGaps, setVerificationGaps] = useState("");
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<{ code: string; message: string }>();
  const requestIdRef = useRef<string>();
  const evidences = useMemo(
    () =>
      candidates
        .filter((candidate) =>
          selectedConclusionIds.has(candidate.conclusionId),
        )
        .map((candidate) => ({
          experimentId: candidate.experimentId,
          conclusionId: candidate.conclusionId,
        })),
    [candidates, selectedConclusionIds],
  );
  const normalizedStatement = statement.trim();
  const normalizedApplicabilityConditions = applicabilityConditions.trim();
  const normalizedVerificationGaps = verificationGaps.trim();
  const validationError = validationMessage(
    evidences.length,
    normalizedStatement,
    normalizedApplicabilityConditions,
    normalizedVerificationGaps,
  );

  const resetRequest = () => {
    requestIdRef.current = undefined;
    setError(undefined);
  };
  const toggleEvidence = (conclusionId: string) => {
    setSelectedConclusionIds((current) => {
      const next = new Set(current);
      if (next.has(conclusionId)) next.delete(conclusionId);
      else next.add(conclusionId);
      return next;
    });
    resetRequest();
  };
  const create = async () => {
    if (validationError || isCreating) return;
    const requestId = requestIdRef.current ?? crypto.randomUUID();
    requestIdRef.current = requestId;
    setIsCreating(true);
    setError(undefined);
    try {
      const response = await createInsight({
        requestId,
        evidences,
        statement: normalizedStatement,
        applicabilityConditions: normalizedApplicabilityConditions,
        verificationGaps: normalizedVerificationGaps,
      });
      if (response.data) {
        onCreated(response.data);
        setIsConfirmOpen(false);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "知見を記録できませんでした。",
        },
      );
    } catch {
      setError({ code: "UNKNOWN", message: "知見を記録できませんでした。" });
    } finally {
      setIsCreating(false);
    }
  };
  const disabled = isCreating || candidates.length < 2;

  return (
    <>
      <Card id="create-insight-form">
        <CardHeader>
          <CardTitle>知見を記録</CardTitle>
        </CardHeader>
        <CardContent className="space-y-5">
          {error && (
            <Alert id="create-insight-error" role="alert" variant="destructive">
              <AlertCircle />
              <AlertTitle>知見を記録できません</AlertTitle>
              <AlertDescription>{error.message}</AlertDescription>
            </Alert>
          )}
          <fieldset className="space-y-3" disabled={disabled}>
            <legend className="font-medium">根拠を選択</legend>
            <p aria-live="polite" className="text-sm text-muted-foreground">
              {evidences.length}
              件選択中。異なる実験から2件以上を選択してください。
            </p>
            {candidates.map((candidate) => (
              <label
                className="flex cursor-pointer items-start gap-3 rounded-md border p-3 has-[:checked]:border-primary"
                key={candidate.conclusionId}
              >
                <input
                  checked={selectedConclusionIds.has(candidate.conclusionId)}
                  onChange={() => toggleEvidence(candidate.conclusionId)}
                  type="checkbox"
                />
                <span className="min-w-0">
                  <span className="block text-sm font-medium">
                    {candidate.experimentId} / {candidate.conclusionId}
                  </span>
                  <span className="block whitespace-pre-wrap break-words text-sm text-muted-foreground">
                    {candidate.conclusion}
                  </span>
                </span>
              </label>
            ))}
          </fieldset>
          <div className="space-y-2">
            <Label htmlFor="insight-statement">知見文</Label>
            <textarea
              className="min-h-28 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              disabled={disabled}
              id="insight-statement"
              onChange={(event) => {
                setStatement(event.target.value);
                resetRequest();
              }}
              value={statement}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="insight-applicability-conditions">適用条件</Label>
            <textarea
              className="min-h-24 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              disabled={disabled}
              id="insight-applicability-conditions"
              onChange={(event) => {
                setApplicabilityConditions(event.target.value);
                resetRequest();
              }}
              value={applicabilityConditions}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="insight-verification-gaps">検証不足</Label>
            <textarea
              className="min-h-24 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              disabled={disabled}
              id="insight-verification-gaps"
              onChange={(event) => {
                setVerificationGaps(event.target.value);
                resetRequest();
              }}
              value={verificationGaps}
            />
          </div>
          {validationError && (
            <p className="text-sm text-destructive">{validationError}</p>
          )}
          <Button
            disabled={Boolean(validationError) || disabled}
            id="open-create-insight-dialog-button"
            onClick={() => setIsConfirmOpen(true)}
            type="button"
          >
            根拠を確認して記録
          </Button>
        </CardContent>
      </Card>
      <AlertDialog
        onOpenChange={(open) => {
          if (!isCreating) setIsConfirmOpen(open);
        }}
        open={isConfirmOpen}
      >
        <AlertDialogContent id="create-insight-dialog">
          <AlertDialogHeader>
            <AlertDialogTitle>知見を記録しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              選択した根拠と入力内容を保存します。既存の実験、比較、実験条件は変更しません。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="text-sm">
            根拠:{" "}
            {evidences.map((evidence) => evidence.experimentId).join(" / ")}
          </div>
          {error && (
            <Alert role="alert" variant="destructive">
              <AlertCircle />
              <AlertTitle>知見を記録できません</AlertTitle>
              <AlertDescription>{error.message}</AlertDescription>
            </Alert>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isCreating}>戻る</AlertDialogCancel>
            <AlertDialogAction
              disabled={isCreating}
              onClick={(event) => {
                event.preventDefault();
                void create();
              }}
            >
              {isCreating ? (
                <>
                  <LoaderCircle className="animate-spin" /> 記録しています…
                </>
              ) : (
                "知見を記録"
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
