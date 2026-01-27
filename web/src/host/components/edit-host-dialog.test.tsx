import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EditHostDialog } from "./edit-host-dialog";
import { updateHost } from "@/host/infrastructure/host-api";
import { vi } from "vitest";

// Mock the API module
vi.mock("@/host/infrastructure/host-api", () => ({
  updateHost: vi.fn(),
}));

// Mock useRouter
const mockRefresh = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    refresh: mockRefresh,
  }),
}));

describe("EditHostDialog", () => {
  const mockHost = {
    id: "1",
    name: "Test Host",
    hostname: "192.168.1.1",
    user: "root",
    port: 22,
    path: "test-host",
    isWorkstation: false,
  };

  const mockOnOpenChange = vi.fn();
  const mockOnSuccess = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the dialog when open is true", () => {
    render(
      <EditHostDialog
        host={mockHost}
        open={true}
        onOpenChange={mockOnOpenChange}
        onSuccess={mockOnSuccess}
      />,
    );

    expect(screen.getByText("Edit Host")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Test Host")).toBeInTheDocument();
    expect(screen.getByDisplayValue("192.168.1.1")).toBeInTheDocument();
    expect(screen.getByDisplayValue("root")).toBeInTheDocument();
    expect(screen.getByDisplayValue("22")).toBeInTheDocument();
    expect(screen.getByDisplayValue("test-host")).toBeInTheDocument();
  });

  it("does not render the dialog when open is false", () => {
    render(
      <EditHostDialog
        host={mockHost}
        open={false}
        onOpenChange={mockOnOpenChange}
        onSuccess={mockOnSuccess}
      />,
    );

    expect(screen.queryByText("Edit Host")).not.toBeInTheDocument();
  });

  it("calls onOpenChange when Cancel button is clicked", async () => {
    const user = userEvent.setup();
    render(
      <EditHostDialog
        host={mockHost}
        open={true}
        onOpenChange={mockOnOpenChange}
        onSuccess={mockOnSuccess}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(mockOnOpenChange).toHaveBeenCalledWith(false);
  });

  it("submits the form and calls updateHost with correct data", async () => {
    const user = userEvent.setup();
    (updateHost as any).mockResolvedValue({});

    render(
      <EditHostDialog
        host={mockHost}
        open={true}
        onOpenChange={mockOnOpenChange}
        onSuccess={mockOnSuccess}
      />,
    );

    // Change some fields
    const nameInput = screen.getByLabelText("Name");
    await user.clear(nameInput);
    await user.type(nameInput, "Updated Host");

    await user.click(screen.getByRole("button", { name: "Save Changes" }));

    await waitFor(() => {
      expect(updateHost).toHaveBeenCalledWith("1", {
        name: "Updated Host",
        hostname: "192.168.1.1",
        user: "root",
        port: 22,
        path: "test-host",
        is_workstation: false,
      });
      expect(mockOnOpenChange).toHaveBeenCalledWith(false);
      expect(mockRefresh).toHaveBeenCalled();
      expect(mockOnSuccess).toHaveBeenCalled();
    });
  });

  it("displays error message when updateHost fails", async () => {
    const user = userEvent.setup();
    (updateHost as any).mockRejectedValue(new Error("API Error"));

    render(
      <EditHostDialog
        host={mockHost}
        open={true}
        onOpenChange={mockOnOpenChange}
        onSuccess={mockOnSuccess}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Save Changes" }));

    await waitFor(() => {
      expect(screen.getByText("API Error")).toBeInTheDocument();
    });

    expect(mockOnOpenChange).not.toHaveBeenCalled();
    expect(mockOnSuccess).not.toHaveBeenCalled();
  });

  it("shows validation errors for empty fields", async () => {
    const user = userEvent.setup();

    render(
      <EditHostDialog
        host={mockHost}
        open={true}
        onOpenChange={mockOnOpenChange}
        onSuccess={mockOnSuccess}
      />,
    );

    // Clear the name field
    const nameInput = screen.getByLabelText("Name");
    await user.clear(nameInput);

    await user.click(screen.getByRole("button", { name: "Save Changes" }));

    await waitFor(() => {
      expect(screen.getByText("Name is required")).toBeInTheDocument();
    });

    expect(updateHost).not.toHaveBeenCalled();
  });
});
