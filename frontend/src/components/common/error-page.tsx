import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button, buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { EmptyState } from "@/components/common/empty-state";
import { AlertTriangle } from "lucide-react";

interface ErrorPageProps {
  statusCode: number;
  title?: string;
  message?: string;
  showRetryButton?: boolean;
  showBackButton?: boolean;
  onRetry?: () => void;
}

function getDefaultTitle(statusCode: number): string {
  switch (statusCode) {
    case 404:
      return "Page Not Found";
    case 403:
      return "Access Denied";
    case 500:
      return "Server Error";
    default:
      return "Something Went Wrong";
  }
}

function getDefaultMessage(statusCode: number): string {
  switch (statusCode) {
    case 404:
      return "The page you're looking for doesn't exist or has been moved.";
    case 403:
      return "You don't have permission to access this page.";
    case 500:
      return "Something went wrong on our end. Please try again in a moment.";
    default:
      return "An unexpected error occurred.";
  }
}

export function ErrorPage({
  statusCode,
  title,
  message,
  showRetryButton,
  showBackButton,
  onRetry,
}: ErrorPageProps) {
  const navigate = useNavigate();
  const [retrying, setRetrying] = useState(false);

  return (
    <div className="flex min-h-[50vh] items-center justify-center p-6">
      <EmptyState
        className="mx-auto max-w-md"
        headingLevel="h1"
        icon={AlertTriangle}
        title={title ?? getDefaultTitle(statusCode)}
        description={message ?? getDefaultMessage(statusCode)}
        action={
          <div className="flex flex-col items-center gap-2">
            <div className="flex flex-wrap justify-center gap-2">
              {showBackButton && (
                <Button variant="outline" onClick={() => navigate(-1)}>
                  Go Back
                </Button>
              )}
              {showRetryButton && onRetry && (
                <Button
                  onClick={() => {
                    setRetrying(true);
                    onRetry();
                  }}
                >
                  Try Again
                </Button>
              )}
              <Link to="/app" className={cn(buttonVariants({ variant: "outline" }))}>
                Home
              </Link>
            </div>
            <p role="status" className="sr-only">
              {retrying ? "Retrying…" : ""}
            </p>
          </div>
        }
      />
    </div>
  );
}

export function Error404Page() {
  return <ErrorPage statusCode={404} showBackButton />;
}

export function Error403Page() {
  return <ErrorPage statusCode={403} showBackButton />;
}
