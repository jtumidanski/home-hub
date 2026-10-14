import { useRef, useState, type DragEvent } from "react";
import { ImagePlus, RefreshCw, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { RECIPE_IMAGE_ACCEPT, validateRecipeImageFile } from "@/lib/recipe-image";
import type { ImageDraft } from "@/lib/hooks/use-recipe-image-draft";

interface RecipeImageFieldProps {
  draft: ImageDraft;
  previewSrc: string | null;
  onStage: (file: File) => void;
  onRemove: () => void;
  onUndo: () => void;
  disabled?: boolean;
}

export function RecipeImageField({ draft, previewSrc, onStage, onRemove, onUndo, disabled = false }: RecipeImageFieldProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState<string | null>(null);
  const [dragging, setDragging] = useState(false);

  const accept = (file: File | undefined) => {
    if (!file) return;
    const message = validateRecipeImageFile(file);
    setError(message);
    if (!message) onStage(file);
  };

  const openPicker = () => inputRef.current?.click();

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setDragging(false);
    if (!disabled) accept(e.dataTransfer.files[0]);
  };

  return (
    <div className="space-y-2">
      <span className="text-sm font-medium leading-none">Photo</span>
      <input
        ref={inputRef}
        type="file"
        accept={RECIPE_IMAGE_ACCEPT}
        aria-label="Choose photo"
        className="hidden"
        disabled={disabled}
        onChange={(e) => {
          accept(e.target.files?.[0]);
          e.target.value = "";
        }}
      />
      {previewSrc ? (
        <div className="flex items-start gap-4">
          <img src={previewSrc} alt="Recipe photo preview" className="h-32 w-48 rounded-md object-cover" />
          <div className="flex flex-col gap-2">
            <Button type="button" variant="outline" size="sm" onClick={openPicker} disabled={disabled}>
              <RefreshCw className="mr-1 h-4 w-4" /> Replace
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="text-destructive"
              onClick={() => {
                setError(null);
                onRemove();
              }}
              disabled={disabled}
            >
              <Trash2 className="mr-1 h-4 w-4" /> Remove
            </Button>
          </div>
        </div>
      ) : (
        <div
          role="button"
          tabIndex={0}
          aria-label="Add a photo"
          onClick={openPicker}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              openPicker();
            }
          }}
          onDragOver={(e) => {
            e.preventDefault();
            setDragging(true);
          }}
          onDragLeave={() => setDragging(false)}
          onDrop={onDrop}
          className={cn(
            "flex cursor-pointer flex-col items-center justify-center gap-2 rounded-md border border-dashed p-6 text-sm text-muted-foreground",
            dragging && "border-primary bg-muted/50",
            error && "border-destructive",
          )}
        >
          <ImagePlus className="h-6 w-6" />
          <span>Drop a photo here, or click to choose one</span>
          <span className="text-xs">JPEG, PNG or WebP, up to 10 MB</span>
        </div>
      )}
      {draft.kind === "removed" && (
        <p className="text-sm text-muted-foreground">
          Photo will be removed when you save.{" "}
          <button type="button" className="text-primary underline" onClick={onUndo}>
            Undo
          </button>
        </p>
      )}
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}
    </div>
  );
}
