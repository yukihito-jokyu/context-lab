export const formatExperimentDateTime = (value: unknown) => {
  if (!value) return "未取得";
  const date = new Date(String(value));
  if (Number.isNaN(date.getTime())) return "未取得";
  return new Intl.DateTimeFormat("ja-JP", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
};
