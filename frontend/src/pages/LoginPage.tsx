import { useProviders } from "@/lib/hooks/api/use-auth";
import { authService } from "@/services/api/auth";
import { buttonVariants } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader } from "@/components/ui/card";
import { StatusBanner } from "@/components/ui/status-banner";
import { Skeleton } from "@/components/ui/skeleton";
import { BrandMark } from "@/components/common/brand-mark";
import { EmptyState } from "@/components/common/empty-state";

export function LoginPage() {
  const { data, isLoading, isError } = useProviders();
  const providers = data?.data ?? [];

  return (
    <div className="flex min-h-dvh items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <BrandMark className="mx-auto mb-2 size-12" />
          <h1 className="font-heading text-2xl tracking-[-0.01em] min-w-0 break-words">Home Hub</h1>
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
                className={buttonVariants({ variant: "outline", size: "lg", className: "w-full" })}
              >
                Sign in with {provider.attributes.displayName}
              </a>
            ))
          )}
          {!isLoading && providers.length === 0 && (
            <EmptyState
              className="py-6"
              title="No login providers configured"
              description="Ask whoever manages this Home Hub to set up a sign-in provider."
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
