import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RecipeImageField } from "../recipe-image-field";

const noop = () => {};

describe("RecipeImageField", () => {
  it("stages a valid picked file", async () => {
    const user = userEvent.setup();
    const onStage = vi.fn();
    render(<RecipeImageField draft={{ kind: "unchanged" }} previewSrc={null} onStage={onStage} onRemove={noop} onUndo={noop} />);
    const file = new File(["x"], "a.jpg", { type: "image/jpeg" });
    await user.upload(screen.getByLabelText("Choose photo"), file);
    expect(onStage).toHaveBeenCalledWith(file);
  });

  it("shows the inline error and stages nothing for an invalid file", () => {
    const onStage = vi.fn();
    render(<RecipeImageField draft={{ kind: "unchanged" }} previewSrc={null} onStage={onStage} onRemove={noop} onUndo={noop} />);
    const gif = new File(["x"], "a.gif", { type: "image/gif" });
    fireEvent.change(screen.getByLabelText("Choose photo"), { target: { files: [gif] } });
    expect(onStage).not.toHaveBeenCalled();
    expect(screen.getByRole("alert")).toHaveTextContent("Use a JPEG, PNG or WebP image under 10 MB.");
  });

  it("stages a dropped file", () => {
    const onStage = vi.fn();
    render(<RecipeImageField draft={{ kind: "unchanged" }} previewSrc={null} onStage={onStage} onRemove={noop} onUndo={noop} />);
    const file = new File(["x"], "a.png", { type: "image/png" });
    fireEvent.drop(screen.getByRole("button", { name: /add a photo/i }), { dataTransfer: { files: [file] } });
    expect(onStage).toHaveBeenCalledWith(file);
  });

  it("shows preview with Replace and Remove", async () => {
    const user = userEvent.setup();
    const onRemove = vi.fn();
    render(<RecipeImageField draft={{ kind: "unchanged" }} previewSrc="/f.jpg" onStage={noop} onRemove={onRemove} onUndo={noop} />);
    expect(screen.getByRole("img", { name: /recipe photo preview/i })).toHaveAttribute("src", "/f.jpg");
    expect(screen.getByRole("button", { name: /replace/i })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /remove/i }));
    expect(onRemove).toHaveBeenCalled();
  });

  it("offers Undo after a staged removal", async () => {
    const user = userEvent.setup();
    const onUndo = vi.fn();
    render(<RecipeImageField draft={{ kind: "removed" }} previewSrc={null} onStage={noop} onRemove={noop} onUndo={onUndo} />);
    await user.click(screen.getByRole("button", { name: /undo/i }));
    expect(onUndo).toHaveBeenCalled();
  });
});
