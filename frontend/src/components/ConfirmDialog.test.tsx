import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialog } from "./ConfirmDialog";

describe("ConfirmDialog", () => {
  it("calls onCancel and not onConfirm when cancelled", async () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();
    render(<ConfirmDialog message="Delete?" onConfirm={onConfirm} onCancel={onCancel} />);
    await userEvent.setup().click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalled();
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it("calls onConfirm when confirmed", async () => {
    const onConfirm = vi.fn();
    render(<ConfirmDialog message="Delete?" onConfirm={onConfirm} onCancel={vi.fn()} />);
    await userEvent.setup().click(screen.getByRole("button", { name: "Confirm" }));
    expect(onConfirm).toHaveBeenCalled();
  });
});
