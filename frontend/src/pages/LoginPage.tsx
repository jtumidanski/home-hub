import { useProviders } from "@/lib/hooks/api/use-auth";
import { authService } from "@/services/api/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusBanner } from "@/components/ui/status-banner";
import { Skeleton } from "@/components/ui/skeleton";

export function LoginPage() {
  const { data, isLoading, isError } = useProviders();
  const providers = data?.data ?? [];

  return (
    <div className="flex min-h-dvh items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="font-heading text-2xl tracking-[-0.01em] min-w-0 break-words">Home Hub</CardTitle>
          <CardDescription>Sign in to manage your household</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {isLoading ? (
            <Skeleton className="h-10 w-full" role="status" aria-label="Loading" />
          ) : isError ? (
            <StatusBanner variant="destructive">
              Failed to load login providers. Try refreshing the page.
            </StatusBanner>
          ) : (
            providers.map((provider) => (
              <a
                key={provider.id}
                href={authService.getLoginUrl(provider.attributes.slug)}
                className="block"
              >
                <Button variant="outline" className="w-full">
                  Sign in with {provider.attributes.displayName}
                </Button>
              </a>
            ))
          )}
          {!isLoading && providers.length === 0 && (
            <p className="text-center text-sm text-muted-foreground">
              No login providers configured
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
