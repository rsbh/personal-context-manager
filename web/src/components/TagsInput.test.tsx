import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderUI } from "../test/render";
import { TagsInput } from "./TagsInput";

describe("TagsInput", () => {
  it("shows tags as chips", () => {
    renderUI(<TagsInput label="Tags" value={["go", "fmt"]} onChange={vi.fn()} />);
    expect(screen.getByText("go")).toBeInTheDocument();
    expect(screen.getByText("fmt")).toBeInTheDocument();
  });

  it("adds a trimmed tag on Enter and clears the input", async () => {
    const onChange = vi.fn();
    renderUI(<TagsInput label="Tags" value={["go"]} onChange={onChange} />);
    const input = screen.getByLabelText("Tags");
    await userEvent.type(input, "  fmt {Enter}");
    expect(onChange).toHaveBeenCalledWith(["go", "fmt"]);
    expect(input).toHaveValue("");
  });

  it("ignores empty and duplicate tags", async () => {
    const onChange = vi.fn();
    renderUI(<TagsInput label="Tags" value={["go"]} onChange={onChange} />);
    const input = screen.getByLabelText("Tags");
    await userEvent.type(input, "   {Enter}go{Enter}");
    expect(onChange).not.toHaveBeenCalled();
  });

  it("removes a tag", async () => {
    const onChange = vi.fn();
    renderUI(<TagsInput label="Tags" value={["go", "fmt"]} onChange={onChange} />);
    await userEvent.click(screen.getByRole("button", { name: "Remove tag go" }));
    expect(onChange).toHaveBeenCalledWith(["fmt"]);
  });
});
