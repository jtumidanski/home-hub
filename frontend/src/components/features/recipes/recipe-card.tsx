import { useNavigate } from "react-router-dom";
import { Clock, Trash2, Pencil } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent, CardAction } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CardActionMenu } from "@/components/common/card-action-menu";
import { PlannerReadyBadge } from "./planner-ready-badge";
import type { RecipeListItem } from "@/types/models/recipe";

interface RecipeCardProps {
  recipe: RecipeListItem;
  onDelete: (id: string) => void;
}

export function RecipeCard({ recipe, onDelete }: RecipeCardProps) {
  const navigate = useNavigate();
  const { id, attributes } = recipe;
  const totalTime = (attributes.prepTimeMinutes ?? 0) + (attributes.cookTimeMinutes ?? 0);

  return (
    <Card
      size="sm"
      className="cursor-pointer"
      onClick={() => navigate(`/app/recipes/${id}`)}
    >
      <CardHeader>
        <CardTitle className="text-base">{attributes.title}</CardTitle>
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
          {totalTime > 0 && (
            <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
              <Clock className="h-3 w-3" />
              {totalTime} min
            </span>
          )}
          {attributes.classification && (
            <Badge variant="secondary" className="text-xs">{attributes.classification}</Badge>
          )}
          {attributes.totalIngredients > 0 && (
            <span className="text-xs text-muted-foreground">
              {attributes.resolvedIngredients}/{attributes.totalIngredients} resolved
            </span>
          )}
          <PlannerReadyBadge ready={attributes.plannerReady} className="text-xs" />
          {attributes.tags.map((tag) => (
            <Badge key={tag} variant="secondary" className="text-xs">
              {tag}
            </Badge>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
