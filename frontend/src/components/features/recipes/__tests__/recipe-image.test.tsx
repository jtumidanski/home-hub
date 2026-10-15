import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
let state: { src: string | null; isError: boolean };
const useSrc = vi.fn((url: string | null) => (url ? state : { src: null, isError: false }));
vi.mock("@/lib/hooks/use-recipe-image-src", () => ({ useRecipeImageSrc: (url: string | null) => useSrc(url) }));

import { RecipeImage } from "../recipe-image";

const image = {
  thumbUrl: "/api/v1/recipes/r1/image?size=thumb&v=i1",
  fullUrl: "/api/v1/recipes/r1/image?size=full&v=i1",
  width: 1600,
  height: 900,
};

describe("RecipeImage", () => {
  beforeEach(() => {
    state = { src: "blob:resolved", isError: false };
    useSrc.mockClear();
  });

  it("card variant renders the lazy thumb rendition", () => {
    render(<RecipeImage image={image} variant="card" alt="Soup" />);
    const img = screen.getByRole("img", { name: "Soup" });
    expect(useSrc).toHaveBeenCalledWith(image.thumbUrl);
    expect(img).toHaveAttribute("src", "blob:resolved");
    expect(img).toHaveAttribute("loading", "lazy");
  });

  it("hero variant renders the eager full rendition", () => {
    render(<RecipeImage image={image} variant="hero" alt="Soup" />);
    const img = screen.getByRole("img", { name: "Soup" });
    expect(useSrc).toHaveBeenCalledWith(image.fullUrl);
    expect(img).toHaveAttribute("src", "blob:resolved");
    expect(img).toHaveAttribute("loading", "eager");
  });

  it("card variant shows the placeholder when there is no image", () => {
    render(<RecipeImage image={null} variant="card" alt="Soup" />);
    expect(screen.getByTestId("recipe-image-placeholder")).toBeInTheDocument();
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });

  it("hero variant renders nothing when there is no image", () => {
    const { container } = render(<RecipeImage image={null} variant="hero" alt="Soup" />);
    expect(container).toBeEmptyDOMElement();
  });

  it("card variant falls back to the placeholder on load error", () => {
    render(<RecipeImage image={image} variant="card" alt="Soup" />);
    fireEvent.error(screen.getByRole("img", { name: "Soup" }));
    expect(screen.getByTestId("recipe-image-placeholder")).toBeInTheDocument();
  });

  it("hero variant renders nothing on load error", () => {
    const { container } = render(<RecipeImage image={image} variant="hero" alt="Soup" />);
    fireEvent.error(screen.getByRole("img", { name: "Soup" }));
    expect(container).toBeEmptyDOMElement();
  });

  it("a new image after an error renders again", () => {
    const { rerender } = render(<RecipeImage image={image} variant="card" alt="Soup" />);
    fireEvent.error(screen.getByRole("img", { name: "Soup" }));
    const next = { ...image, thumbUrl: image.thumbUrl.replace("v=i1", "v=i2") };
    state = { src: "blob:next", isError: false };
    rerender(<RecipeImage image={next} variant="card" alt="Soup" />);
    expect(screen.getByRole("img", { name: "Soup" })).toHaveAttribute("src", "blob:next");
  });

  it("card variant keeps the placeholder while the image is loading", () => {
    state = { src: null, isError: false };
    render(<RecipeImage image={image} variant="card" alt="Soup" />);
    expect(screen.getByTestId("recipe-image-placeholder")).toBeInTheDocument();
  });

  it("hero variant renders nothing while loading", () => {
    state = { src: null, isError: false };
    const { container } = render(<RecipeImage image={image} variant="hero" alt="Soup" />);
    expect(container).toBeEmptyDOMElement();
  });

  it("falls back like a load error when the fetch fails", () => {
    state = { src: null, isError: true };
    const card = render(<RecipeImage image={image} variant="card" alt="Soup" />);
    expect(screen.getByTestId("recipe-image-placeholder")).toBeInTheDocument();
    card.unmount();
    const { container } = render(<RecipeImage image={image} variant="hero" alt="Soup" />);
    expect(container).toBeEmptyDOMElement();
  });
});
