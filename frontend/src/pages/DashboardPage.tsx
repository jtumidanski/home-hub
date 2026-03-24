import { useAuth } from "@/components/providers/auth-provider";

export function DashboardPage() {
  const { appContext } = useAuth();

  return (
    <div className="p-6">
      <h1 className="text-2xl font-semibold">Dashboard</h1>
      <p className="mt-2 text-muted-foreground">
        Welcome to Home Hub
        {appContext?.attributes.resolvedRole && (
          <> &mdash; you are {appContext.attributes.resolvedRole}</>
        )}
      </p>
    </div>
  );
}
