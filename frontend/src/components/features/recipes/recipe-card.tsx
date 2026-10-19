import { Link, useNavigate } from "react-router-dom";
import { Clock, Trash2, Pencil, CheckCircle2, AlertCircle, Circle } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent, CardAction } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CardActionMenu } from "@/components/common/card-action-menu";
import { filterClassificationTags } from "@/lib/constants/recipe";
import { RecipeImage } from "./recipe-image";
import { toTitleCase } from "@/lib/utils";
import type { RecipeListItem, RecipeListAttributes } from "@/types/models/recipe";

interface RecipeCardProps {
  recipe: RecipeListItem;
  onDelete: (id: string) => void;
}

function RecipeStatusIcon({ attrs }: { attrs: RecipeListAttributes }) {
  const allResolved = attrs.totalIngredients > 0 && attrs.resolvedIngredients === attrs.totalIngredients;
  const hasUnresolved = attrs.totalIngredients > 0 && attrs.resolvedIngredients < attrs.totalIngredients;

  // Status text: planner status + any issues (no classification, no resolution count)
  const parts: string[] = [];
  if (attrs.plannerReady) {
    parts.push("Planner Ready");
  } else {
    parts.push("Not Planner Ready");
  }
  if (hasUnresolved) {
    parts.push("has unresolved ingredients");
  }

  const statusText = parts.join(" — ");

  if (attrs.plannerReady && allResolved) {
    return (
      <span className="inline-flex">
        <CheckCircle2 aria-hidden="true" className="h-4 w-4 text-success shrink-0" />
        <span className="sr-only">{statusText}</span>
      </span>
    );
  }
  if (attrs.classification || attrs.totalIngredients > 0) {
    return (
      <span className="inline-flex">
        <AlertCircle aria-hidden="true" className="h-4 w-4 text-warning-text shrink-0" />
        <span className="sr-only">{statusText}</span>
      </span>
    );
  }
  return (
    <span className="inline-flex">
      <Circle aria-hidden="true" className="h-4 w-4 text-muted-foreground shrink-0" />
      <span className="sr-only">{statusText}</span>
    </span>
  );
}

export function RecipeCard({ recipe, onDelete }: RecipeCardProps) {
  const navigate = useNavigate();
  const { id, attributes } = recipe;
  const totalTime = (attributes.prepTimeMinutes ?? 0) + (attributes.cookTimeMinutes ?? 0);
  const nonClassTags = filterClassificationTags(attributes.tags);

  // Show classification as a tag badge if present
  const allTags = [
    ...(attributes.classification ? [attributes.classification] : []),
    ...nonClassTags,
  ];

  return (
    <Card
      size="sm"
      className="relative md:pt-0 focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-ring"
    >
      <div className="flex items-start gap-3 pl-3 md:flex-col md:items-stretch md:pl-0">
        <RecipeImage
          image={attributes.image}
          variant="card"
          alt={attributes.title}
          className="size-16 shrink-0 rounded-md md:aspect-video md:h-auto md:w-full md:rounded-none"
        />
        <div className="flex min-w-0 flex-1 flex-col gap-3">
          <CardHeader className="max-md:pl-0">
            <div className="flex items-center gap-2">
              <RecipeStatusIcon attrs={attributes} />
              <CardTitle className="text-base">
                <Link
                  to={`/app/recipes/${id}`}
                  className="outline-none after:absolute after:inset-0 after:content-['']"
                >
                  {attributes.title}
                </Link>
              </CardTitle>
            </div>
            <CardAction className="relative z-10">
              <CardActionMenu
                actions={[
                  {
                    icon: <Pencil className="h-4 w-4" />,
                    label: "Edit",
                    onClick: () => navigate(`/app/recipes/${id}/edit`),
                  },
                  {
                    icon: <Trash2 className="h-4 w-4" />,
                    label: "Delete",
                    onClick: () => onDelete(id),
                    variant: "destructive",
                  },
                ]}
              />
            </CardAction>
          </CardHeader>
          <CardContent className="max-md:pl-0">
            {attributes.description && (
              <p className="text-sm text-muted-foreground line-clamp-2 mb-2">
                {attributes.description}
              </p>
            )}
            <div className="flex flex-wrap items-center gap-2">
              <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
                <Clock className="h-3 w-3" />
                {totalTime > 0 ? `${totalTime} min` : "??? min"}
              </span>
              {(attributes.usageCount ?? 0) > 0 && (
                <span className="text-xs text-muted-foreground">
                  cooked {attributes.usageCount}x
                </span>
              )}
              {allTags.map((tag) => (
                <Badge key={tag} variant="secondary" className="text-xs">
                  {toTitleCase(tag)}
                </Badge>
              ))}
            </div>
          </CardContent>
        </div>
      </div>
    </Card>
  );
}
