import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderUI } from "../test/render";
import { DeleteDialog } from "./DeleteDialog";

interface Setup {
  kind?: "section" | "project";
  notEmpty?: boolean;
}

function renderDialog({ kind = "section", notEmpty = false }: Setup = {}) {
  const onDelete = vi.fn<(force: boolean) => void>();
  const onOpenChange = vi.fn<(open: boolean) => void>();
  renderUI(
    <DeleteDialog
      kind={kind}
      name="style"
      open
      onOpenChange={onOpenChange}
      notEmpty={notEmpty}
      pending={false}
      onDelete={onDelete}
    />,
  );
  return { onDelete, onOpenChange };
}

describe("DeleteDialog", () => {
  it("confirms a normal delete", async () => {
    const { onDelete } = renderDialog();
    const dialog = screen.getByRole("alertdialog", { name: "Delete section?" });
    expect(dialog).toHaveTextContent("style");
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    expect(onDelete).toHaveBeenCalledWith(false);
  });

  it("asks to delete all when not empty", async () => {
    const { onDelete } = renderDialog({ notEmpty: true });
    expect(screen.getByRole("alertdialog", { name: "Section is not empty" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Delete all" }));
    expect(onDelete).toHaveBeenCalledWith(true);
  });

  it("names a project", () => {
    renderDialog({ kind: "project", notEmpty: true });
    expect(screen.getByRole("alertdialog", { name: "Project is not empty" })).toBeInTheDocument();
  });

  it("closes on Cancel", async () => {
    const { onDelete, onOpenChange } = renderDialog();
    await userEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onDelete).not.toHaveBeenCalled();
  });
});
