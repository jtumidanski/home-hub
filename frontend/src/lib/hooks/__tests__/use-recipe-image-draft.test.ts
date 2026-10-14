import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useRecipeImageDraft } from "../use-recipe-image-draft";

const existing = { thumbUrl: "/t.jpg", fullUrl: "/f.jpg", width: 10, height: 10 };
const file = () => new File(["x"], "a.jpg", { type: "image/jpeg" });

describe("useRecipeImageDraft", () => {
  let n = 0;
  beforeEach(() => {
    n = 0;
    URL.createObjectURL = vi.fn(() => `blob:${++n}`);
    URL.revokeObjectURL = vi.fn();
  });

  it("starts unchanged and previews the existing image", () => {
    const { result } = renderHook(() => useRecipeImageDraft(existing));
    expect(result.current.draft.kind).toBe("unchanged");
    expect(result.current.previewSrc).toBe("/f.jpg");
  });

  it("stage creates an object URL; replacing revokes the previous one", () => {
    const { result } = renderHook(() => useRecipeImageDraft(null));
    act(() => result.current.stage(file()));
    expect(result.current.previewSrc).toBe("blob:1");
    act(() => result.current.stage(file()));
    expect(result.current.previewSrc).toBe("blob:2");
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:1");
  });

  it("remove with an existing image stages removal; undo restores", () => {
    const { result } = renderHook(() => useRecipeImageDraft(existing));
    act(() => result.current.remove());
    expect(result.current.draft.kind).toBe("removed");
    expect(result.current.previewSrc).toBeNull();
    act(() => result.current.undo());
    expect(result.current.draft.kind).toBe("unchanged");
    expect(result.current.previewSrc).toBe("/f.jpg");
  });

  it("remove of a staged file without an existing image returns to unchanged", () => {
    const { result } = renderHook(() => useRecipeImageDraft(null));
    act(() => result.current.stage(file()));
    act(() => result.current.remove());
    expect(result.current.draft.kind).toBe("unchanged");
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:1");
  });

  it("revokes the object URL on unmount", () => {
    const { result, unmount } = renderHook(() => useRecipeImageDraft(null));
    act(() => result.current.stage(file()));
    unmount();
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:1");
  });
});
