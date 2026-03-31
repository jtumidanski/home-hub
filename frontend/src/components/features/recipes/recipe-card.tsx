import { useNavigate } from "react-router-dom";
import { Clock, Trash2, Pencil, CheckCircle2, AlertCircle, Circle } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent, CardAction } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CardActionMenu } from "@/components/common/card-action-menu";
import { filterClassificationTags } from "@/lib/constants/recipe";
import { toTitleCase } from "@/lib/utils";
import type { RecipeListItem, RecipeListAttributes } from "@/types/models/recipe";

interface RecipeCardProps {
  recipe: RecipeListItem;
  onDelete: (id: string) => void;
}

function RecipeStatusIcon({ attrs }: { attrs: RecipeListAttributes }) {
  const allResolved = attrs.totalIngredients > 0 && attrs.resolvedIngredients === attrs.totalIngredients;
  const hasUnresolved = attrs.totalIngredients > 0 && attrs.resolvedIngredients < attrs.totalIngredients;

  // Tooltip: planner status + any issues (no classification, no resolution count)
  const parts: string[] = [];
  if (attrs.plannerReady) {
    parts.push("Planner Ready");
  } else {
    parts.push("Not Planner Ready");
  }
  if (hasUnresolved) {
    parts.push("has unresolved ingredients");
  }

  const tooltip = parts.join(" — ");

  if (attrs.plannerReady && allResolved) {
    return <span title={tooltip}><CheckCircle2 className="h-4 w-4 text-green-500 shrink-0" /></span>;
  }
  if (attrs.classification || attrs.totalIngredients > 0) {
    return <span title={tooltip}><AlertCircle className="h-4 w-4 text-yellow-500 shrink-0" /></span>;
  }
  return <span title={tooltip}><Circle className="h-4 w-4 text-muted-foreground/40 shrink-0" /></span>;
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
      className="cursor-pointer"
      onClick={() => navigate(`/app/recipes/${id}`)}
    >
      <CardHeader>
        <div className="flex items-center gap-2">
          <RecipeStatusIcon attrs={attributes} />
          <CardTitle className="text-base">{attributes.title}</CardTitle>
        </div>
        <CardAction>
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
      <CardContent>
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
          {allTags.map((tag) => (
            <Badge key={tag} variant="secondary" className="text-xs">
              {toTitleCase(tag)}
            </Badge>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
