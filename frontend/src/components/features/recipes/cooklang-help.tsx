import { useState } from "react";
import { HelpCircle, ChevronDown, ChevronUp } from "lucide-react";
import { Button } from "@/components/ui/button";

export function CooklangHelp() {
  const [open, setOpen] = useState(false);

  return (
    <div>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        className="gap-1 text-xs text-muted-foreground"
        onClick={() => setOpen(!open)}
      >
        <HelpCircle className="h-3 w-3" />
        Cooklang syntax
        {open ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
      </Button>

      {open && (
        <div className="mt-2 rounded-md border bg-muted/50 p-3 text-xs space-y-3">
          <div className="space-y-1.5">
            <p className="font-semibold">Metadata (between --- markers):</p>
            <p><code className="bg-muted px-1 rounded">tags: italian, pasta</code> or <code className="bg-muted px-1 rounded">tags: [italian, pasta]</code></p>
            <p><code className="bg-muted px-1 rounded">source: https://...</code> — recipe source URL</p>
            <p><code className="bg-muted px-1 rounded">servings: 4</code></p>
            <p><code className="bg-muted px-1 rounded">prep time: 20 minutes</code></p>
            <p><code className="bg-muted px-1 rounded">cook time: 35 minutes</code></p>
          </div>
          <div className="space-y-1.5">
            <p className="font-semibold">Recipe syntax:</p>
            <p><code className="bg-muted px-1 rounded">@ingredient{"{qty%unit}"}</code> — e.g. <code>@salt{"{1%tsp}"}</code></p>
            <p><code className="bg-muted px-1 rounded">@ingredient{"{qty}"}</code> — e.g. <code>@eggs{"{3}"}</code></p>
            <p><code className="bg-muted px-1 rounded">@ingredient</code> — single word, no quantity</p>
            <p><code className="bg-muted px-1 rounded">#cookware{"{}"}</code> — e.g. <code>#large pot{"{}"}</code></p>
            <p><code className="bg-muted px-1 rounded">~{"{qty%unit}"}</code> — timer, e.g. <code>~{"{8%minutes}"}</code></p>
          </div>
          <div className="space-y-1.5">
            <p className="font-semibold">Structure:</p>
            <p><code className="bg-muted px-1 rounded">= Section Name</code> — section header</p>
            <p><code className="bg-muted px-1 rounded">&gt; Note text</code> — note (not a step)</p>
            <p><code className="bg-muted px-1 rounded">--</code> — line comment</p>
            <p>Blank lines separate steps.</p>
          </div>
        </div>
      )}
    </div>
  );
}
