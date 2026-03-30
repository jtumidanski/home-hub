import { useState } from "react";
import { Copy, Download, FileText } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useExportMarkdown } from "@/lib/hooks/api/use-meals";

interface ExportModalProps {
  open: boolean;
  onClose: () => void;
  planId: string;
  planName: string;
}

export function ExportModal({ open, onClose, planId, planName }: ExportModalProps) {
  const [markdown, setMarkdown] = useState<string | null>(null);
  const [showRaw, setShowRaw] = useState(false);
  const exportMutation = useExportMarkdown();

  const handleOpen = () => {
    if (!markdown) {
      exportMutation.mutate(planId, {
        onSuccess: (data) => setMarkdown(data),
      });
    }
  };

  const copyToClipboard = async () => {
    if (!markdown) return;
    await navigator.clipboard.writeText(markdown);
    toast.success("Copied to clipboard");
  };

  const downloadFile = () => {
    if (!markdown) return;
    const blob = new Blob([markdown], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${planName.replace(/\s+/g, "_")}.md`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (o) handleOpen();
        if (!o) {
          onClose();
          setMarkdown(null);
          setShowRaw(false);
        }
      }}
    >
      <DialogContent className="sm:max-w-[600px] max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>Export Plan</DialogTitle>
        </DialogHeader>

        <div className="flex gap-2 mb-2">
          <Button variant="outline" size="sm" onClick={copyToClipboard} disabled={!markdown}>
            <Copy className="h-4 w-4 mr-1" /> Copy
          </Button>
          <Button variant="outline" size="sm" onClick={downloadFile} disabled={!markdown}>
            <Download className="h-4 w-4 mr-1" /> Download .md
          </Button>
          <Button
            variant={showRaw ? "default" : "outline"}
            size="sm"
            onClick={() => setShowRaw(!showRaw)}
            disabled={!markdown}
          >
            <FileText className="h-4 w-4 mr-1" /> Raw
          </Button>
        </div>

        <div className="flex-1 overflow-y-auto border rounded p-4 bg-muted/30 text-sm">
          {exportMutation.isPending ? (
            <p className="text-muted-foreground">Generating export...</p>
          ) : markdown ? (
            showRaw ? (
              <pre className="whitespace-pre-wrap font-mono text-xs">{markdown}</pre>
            ) : (
              <div className="prose prose-sm dark:prose-invert max-w-none">
                <MarkdownPreview content={markdown} />
              </div>
            )
          ) : (
            <p className="text-muted-foreground">Loading...</p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

function parseBoldText(text: string): React.ReactNode[] {
  const parts = text.split(/\*\*(.+?)\*\*/g);
  return parts.map((part, i) =>
    i % 2 === 1 ? <strong key={i}>{part}</strong> : part
  );
}

function MarkdownPreview({ content }: { content: string }) {
  const lines = content.split("\n");
  return (
    <div>
      {lines.map((line, i) => {
        if (line.startsWith("# ")) return <h1 key={i} className="text-lg font-bold mt-2">{line.slice(2)}</h1>;
        if (line.startsWith("## ")) return <h2 key={i} className="text-base font-semibold mt-3 mb-1">{line.slice(3)}</h2>;
        if (line.startsWith("- ")) {
          const text = line.slice(2);
          return (
            <div key={i} className="ml-4 before:content-['•'] before:mr-2 before:text-muted-foreground">
              <span>{parseBoldText(text)}</span>
            </div>
          );
        }
        if (line.trim() === "") return <div key={i} className="h-2" />;
        return <p key={i}>{line}</p>;
      })}
    </div>
  );
}
