import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Link } from "react-router";
import { describe, expect, it } from "vitest";
import { renderUI } from "../test/render";
import { UnsavedGuard } from "./UnsavedGuard";

interface PageProps {
  dirty: boolean;
}

function Page({ dirty }: PageProps) {
  return (
    <>
      <UnsavedGuard when={dirty} />
      <Link to="/other">Other</Link>
    </>
  );
}

describe("UnsavedGuard", () => {
  it("navigates without a dialog when clean", async () => {
    renderUI(<Page dirty={false} />, "/edit/x", "/edit/*");
    await userEvent.click(screen.getByRole("link", { name: "Other" }));
    expect(screen.getByLabelText("Location")).toHaveTextContent("/other");
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
  });

  it("asks before leaving when dirty: Stay keeps the page", async () => {
    renderUI(<Page dirty />, "/edit/x", "/edit/*");
    await userEvent.click(screen.getByRole("link", { name: "Other" }));
    expect(await screen.findByRole("alertdialog", { name: "Unsaved changes" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Stay" }));
    expect(screen.getByLabelText("Location")).toHaveTextContent("/edit/x");
  });

  it("asks before leaving when dirty: Leave navigates", async () => {
    renderUI(<Page dirty />, "/edit/x", "/edit/*");
    await userEvent.click(screen.getByRole("link", { name: "Other" }));
    await userEvent.click(await screen.findByRole("button", { name: "Leave" }));
    expect(screen.getByLabelText("Location")).toHaveTextContent("/other");
  });

  it("blocks tab close when dirty", () => {
    renderUI(<Page dirty />, "/edit/x", "/edit/*");
    const event = new Event("beforeunload", { cancelable: true });
    window.dispatchEvent(event);
    expect(event.defaultPrevented).toBe(true);
  });

  it("does not block tab close when clean", () => {
    renderUI(<Page dirty={false} />, "/edit/x", "/edit/*");
    const event = new Event("beforeunload", { cancelable: true });
    window.dispatchEvent(event);
    expect(event.defaultPrevented).toBe(false);
  });
});
