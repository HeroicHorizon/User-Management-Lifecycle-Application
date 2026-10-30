import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "../api/client";
import { renderWithProviders } from "../testUtils";
import { UserForm } from "./UserForm";

afterEach(() => vi.unstubAllGlobals());

async function fillValid(user: ReturnType<typeof userEvent.setup>, email = "john@example.com") {
  await user.type(screen.getByLabelText("First Name"), "John");
  await user.type(screen.getByLabelText("Last Name"), "Doe");
  await user.type(screen.getByLabelText("Email Address"), email);
  await user.type(screen.getByLabelText("Phone Number"), "+911234567890");
  await user.type(screen.getByLabelText("Department"), "Engineering");
  await user.type(screen.getByLabelText("Designation"), "Engineer");
}

describe("UserForm", () => {
  it("shows required-field errors and does not submit", async () => {
    const onSubmit = vi.fn();
    renderWithProviders(<UserForm submitLabel="Create" onSubmit={onSubmit} />);
    await userEvent.setup().click(screen.getByRole("button", { name: "Create" }));
    expect(screen.getAllByText("This field is required.")).toHaveLength(5);
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("rejects invalid email", async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<UserForm submitLabel="Create" onSubmit={onSubmit} />);
    await fillValid(user, "not-an-email");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(screen.getByText("Enter a valid email address.")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("submits valid data", async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    const user = userEvent.setup();
    renderWithProviders(<UserForm submitLabel="Create" onSubmit={onSubmit} />);
    await fillValid(user);
    await user.click(screen.getByRole("button", { name: "Create" }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith({
      firstName: "John",
      lastName: "Doe",
      email: "john@example.com",
      phone: "+911234567890",
      department: "Engineering",
      designation: "Engineer",
    });
  });

  it('shows "A user with this email already exists." on a 409 response', async () => {
    const onSubmit = vi.fn().mockRejectedValue(new ApiError("dup", 409, "DUPLICATE_EMAIL"));
    const user = userEvent.setup();
    renderWithProviders(<UserForm submitLabel="Create" onSubmit={onSubmit} />);
    await fillValid(user);
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(await screen.findByText("A user with this email already exists.")).toBeInTheDocument();
  });

  it("shows an error notification for other failures", async () => {
    const onSubmit = vi
      .fn()
      .mockRejectedValue(new ApiError("Unable to reach the server. Please try again later.", 0, "NETWORK"));
    const user = userEvent.setup();
    renderWithProviders(<UserForm submitLabel="Create" onSubmit={onSubmit} />);
    await fillValid(user);
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(await screen.findByText("Unable to reach the server. Please try again later.")).toBeInTheDocument();
  });
});
