import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
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
      return "An unexpected error occurred. Please try again later.";
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

  return (
    <div className="flex min-h-[50vh] items-center justify-center p-6">
      <EmptyState
        className="mx-auto max-w-md"
        icon={AlertTriangle}
        title={title ?? getDefaultTitle(statusCode)}
        description={message ?? getDefaultMessage(statusCode)}
        action={
          showBackButton || (showRetryButton && onRetry) ? (
            <div className="flex justify-center gap-2">
              {showBackButton && (
                <Button variant="outline" onClick={() => navigate(-1)}>
                  Go Back
                </Button>
              )}
              {showRetryButton && onRetry && (
                <Button onClick={onRetry}>Try Again</Button>
              )}
            </div>
          ) : undefined
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
