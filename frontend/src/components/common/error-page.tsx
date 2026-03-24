import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
      <Card className="mx-auto max-w-md text-center">
        <CardHeader>
          <div className="mx-auto mb-2">
            <AlertTriangle className="h-10 w-10 text-muted-foreground" />
          </div>
          <CardTitle>{title ?? getDefaultTitle(statusCode)}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            {message ?? getDefaultMessage(statusCode)}
          </p>
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
        </CardContent>
      </Card>
    </div>
  );
}

export function Error404Page() {
  return <ErrorPage statusCode={404} showBackButton />;
}

export function Error403Page() {
  return <ErrorPage statusCode={403} showBackButton />;
}
