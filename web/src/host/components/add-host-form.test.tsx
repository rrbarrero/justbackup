import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AddHostForm } from "./add-host-form";
import { createHost } from "@/host/infrastructure/host-api";
import { vi } from "vitest";

// Mock the API module
vi.mock("@/host/infrastructure/host-api", () => ({
  createHost: vi.fn(),
}));

// Mock useRouter
const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

describe("AddHostForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the form with all fields", () => {
    render(<AddHostForm />);

    expect(screen.getByText("Add New Host")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
    expect(screen.getByLabelText("Host Path (Slug)")).toBeInTheDocument();
    expect(screen.getByLabelText("Hostname / IP")).toBeInTheDocument();
    expect(screen.getByLabelText("SSH User")).toBeInTheDocument();
    expect(screen.getByLabelText("SSH Port")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Add Host" }),
    ).toBeInTheDocument();
  });

  it("auto-generates slug from name", async () => {
    const user = userEvent.setup();
    render(<AddHostForm />);

    const nameInput = screen.getByLabelText("Name");
    const pathInput = screen.getByLabelText("Host Path (Slug)");

    await user.type(nameInput, "My Test Server");

    await waitFor(() => {
      expect(pathInput).toHaveValue("my-test-server");
    });
  });

  it("submits the form successfully and redirects", async () => {
    const user = userEvent.setup();
    (createHost as any).mockResolvedValue({});

    render(<AddHostForm />);

    await user.type(screen.getByLabelText("Name"), "Production Server");
    await user.type(screen.getByLabelText("Hostname / IP"), "192.168.1.100");
    await user.type(screen.getByLabelText("SSH User"), "admin");
    await user.clear(screen.getByLabelText("SSH Port"));
    await user.type(screen.getByLabelText("SSH Port"), "2222");

    await user.click(screen.getByRole("button", { name: "Add Host" }));

    await waitFor(() => {
      expect(createHost).toHaveBeenCalledWith({
        name: "Production Server",
        hostname: "192.168.1.100",
        user: "admin",
        port: 2222,
        path: "production-server",
        is_workstation: false,
      });
      expect(mockPush).toHaveBeenCalledWith("/hosts");
    });
  });

  it("displays error message when createHost fails", async () => {
    const user = userEvent.setup();
    (createHost as any).mockRejectedValue(new Error("Network error"));

    render(<AddHostForm />);

    await user.type(screen.getByLabelText("Name"), "Test Host");
    await user.type(screen.getByLabelText("Hostname / IP"), "192.168.1.1");
    await user.type(screen.getByLabelText("SSH User"), "root");

    await user.click(screen.getByRole("button", { name: "Add Host" }));

    await waitFor(() => {
      expect(screen.getByText("Network error")).toBeInTheDocument();
    });

    expect(mockPush).not.toHaveBeenCalled();
  });

  it("shows validation errors for empty required fields", async () => {
    const user = userEvent.setup();

    render(<AddHostForm />);

    await user.click(screen.getByRole("button", { name: "Add Host" }));

    await waitFor(() => {
      expect(screen.getByText("Name is required")).toBeInTheDocument();
      expect(screen.getByText("Hostname is required")).toBeInTheDocument();
      expect(screen.getByText("User is required")).toBeInTheDocument();
      expect(screen.getByText("Path is required")).toBeInTheDocument();
    });

    expect(createHost).not.toHaveBeenCalled();
  });

  it("disables submit button while loading", async () => {
    const user = userEvent.setup();
    (createHost as any).mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100)),
    );

    render(<AddHostForm />);

    await user.type(screen.getByLabelText("Name"), "Test");
    await user.type(screen.getByLabelText("Hostname / IP"), "192.168.1.1");
    await user.type(screen.getByLabelText("SSH User"), "root");

    const submitButton = screen.getByRole("button", { name: "Add Host" });
    await user.click(submitButton);

    expect(submitButton).toBeDisabled();
    expect(screen.getByText("Adding Host...")).toBeInTheDocument();
  });
});
