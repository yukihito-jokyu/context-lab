import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function App() {
  return (
    <main className="min-h-screen bg-background p-8 text-foreground">
      <Card className="mx-auto max-w-3xl">
        <CardHeader>
          <CardTitle>Context Lab</CardTitle>
        </CardHeader>
        <CardContent>
          Wails関数単位Issueごとに実験画面を実装します。
        </CardContent>
      </Card>
    </main>
  );
}
