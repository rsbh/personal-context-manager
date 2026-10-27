import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderUI } from "../test/render";
import { NameDialog } from "./NameDialog";

function renderDialog(onSubmit = vi.fn<(name: string) => void>()) {
  renderUI(
    <NameDialog
      title="New section"
      description="In api"
      open
      onOpenChange={vi.fn<(open: boolean) => void>()}
      pending={false}
      onSubmit={onSubmit}
    />,
  );
  return onSubmit;
}

describe("NameDialog", () => {
  it("shows title and description", () => {
    renderDialog();
    expect(screen.getByRole("dialog", { name: "New section" })).toHaveTextContent("In api");
  });

  it("disables Create without a name", async () => {
    renderDialog();
    expect(screen.getByRole("button", { name: "Create" })).toBeDisabled();
    await userEvent.type(screen.getByLabelText("Name"), "   ");
    expect(screen.getByRole("button", { name: "Create" })).toBeDisabled();
  });

  it("submits the trimmed name with Create", async () => {
    const onSubmit = renderDialog();
    await userEvent.type(screen.getByLabelText("Name"), " style ");
    await userEvent.click(screen.getByRole("button", { name: "Create" }));
    expect(onSubmit).toHaveBeenCalledWith("style");
  });

  it("submits with Enter", async () => {
    const onSubmit = renderDialog();
    await userEvent.type(screen.getByLabelText("Name"), "style{Enter}");
    expect(onSubmit).toHaveBeenCalledWith("style");
  });
});
