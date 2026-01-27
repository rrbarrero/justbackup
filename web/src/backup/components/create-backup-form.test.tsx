import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CreateBackupForm } from "./create-backup-form";
import {
  createBackup,
  measureSize,
  getTaskResult,
} from "@/backup/infrastructure/backup-api";
import { vi } from "vitest";

// Mock the API module
vi.mock("@/backup/infrastructure/backup-api", () => ({
  createBackup: vi.fn(),
  measureSize: vi.fn(),
  getTaskResult: vi.fn(),
}));

// Mock useRouter
const mockPush = vi.fn();
const mockBack = vi.fn();
const mockRefresh = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
    back: mockBack,
    refresh: mockRefresh,
  }),
}));

describe("CreateBackupForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Rendering", () => {
    it("renders the form with all fields", async () => {
      const user = userEvent.setup();
      render(<CreateBackupForm hostId="test-host-id" />);

      expect(screen.getByText("Create New Backup Task")).toBeInTheDocument();
      expect(screen.getByLabelText("Source Path")).toBeInTheDocument();
      expect(screen.getByLabelText("Destination Name")).toBeInTheDocument();
      expect(screen.getByLabelText("Schedule (Cron)")).toBeInTheDocument();
      expect(screen.getByLabelText("Exclude Patterns")).toBeInTheDocument();

      // Check Advanced tab content
      await user.click(screen.getByText("Advanced Settings"));
      expect(screen.getByLabelText(/incremental backup/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/encrypt backup/i)).toBeInTheDocument();

      expect(
        screen.getByRole("button", { name: "Create Backup Task" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Cancel" }),
      ).toBeInTheDocument();
    });

    it("has default values for fields", async () => {
      const user = userEvent.setup();
      render(<CreateBackupForm hostId="test-host-id" />);

      await user.click(screen.getByText("Advanced Settings"));
      const incrementalSwitch = screen.getByRole("switch", {
        name: /incremental backup/i,
      });
      expect(incrementalSwitch).not.toBeChecked();

      const encryptedSwitch = screen.getByRole("switch", {
        name: /encrypt backup/i,
      });
      expect(encryptedSwitch).not.toBeChecked();
    });
  });

  describe("Form Validation", () => {
    it("shows validation errors for empty required fields", async () => {
      const user = userEvent.setup();
      render(<CreateBackupForm hostId="test-host-id" />);

      await user.click(
        screen.getByRole("button", { name: "Create Backup Task" }),
      );

      await waitFor(() => {
        expect(screen.getByText("Path is required")).toBeInTheDocument();
        expect(screen.getByText("Destination is required")).toBeInTheDocument();
      });

      expect(createBackup).not.toHaveBeenCalled();
    });

    it("shows validation error for invalid cron expression", async () => {
      const user = userEvent.setup();
      render(<CreateBackupForm hostId="test-host-id" />);

      const scheduleInput = screen.getByLabelText("Schedule (Cron)");
      await user.clear(scheduleInput);
      await user.type(scheduleInput, "invalid cron");
      await user.click(
        screen.getByRole("button", { name: "Create Backup Task" }),
      );

      await waitFor(() => {
        expect(
          screen.getByText(
            "Invalid cron expression (e.g. '0 0 * * *' or '@daily')",
          ),
        ).toBeInTheDocument();
      });
    });
  });

  describe("User Interactions", () => {
    it("allows typing in text fields", async () => {
      const user = userEvent.setup();
      render(<CreateBackupForm hostId="test-host-id" />);

      const destInput = screen.getByLabelText("Destination Name");
      const excludesInput = screen.getByLabelText("Exclude Patterns");

      await user.type(destInput, "my-backup");
      await user.type(excludesInput, "*.log,tmp/");

      expect(destInput).toHaveValue("my-backup");
      expect(excludesInput).toHaveValue("*.log,tmp/");
    });

    it("toggles incremental backup switch", async () => {
      const user = userEvent.setup();
      render(<CreateBackupForm hostId="test-host-id" />);

      await user.click(screen.getByText("Advanced Settings"));
      const incrementalSwitch = screen.getByRole("switch", {
        name: /incremental backup/i,
      });
      expect(incrementalSwitch).not.toBeChecked();

      await user.click(incrementalSwitch);
      expect(incrementalSwitch).toBeChecked();

      await user.click(incrementalSwitch);
      expect(incrementalSwitch).not.toBeChecked();
    });

    it("calls router.back() when Cancel button is clicked", async () => {
      const user = userEvent.setup();
      render(<CreateBackupForm hostId="test-host-id" />);

      await user.click(screen.getByRole("button", { name: "Cancel" }));

      expect(mockBack).toHaveBeenCalled();
    });
  });

  describe("Form Submission", () => {
    it("submits form successfully with valid data", async () => {
      const user = userEvent.setup();
      (createBackup as any).mockResolvedValue({});

      render(<CreateBackupForm hostId="test-host-id" />);

      const destInput = screen.getByLabelText("Destination Name");
      await user.type(destInput, "backup-dest");

      const pathInput = screen.getByLabelText("Source Path");
      await user.type(pathInput, "/var/www");

      const excludesInput = screen.getByLabelText("Exclude Patterns");
      await user.type(excludesInput, "*.log, tmp/");

      await user.click(
        screen.getByRole("button", { name: "Create Backup Task" }),
      );

      await waitFor(() => {
        expect(createBackup).toHaveBeenCalledWith({
          host_id: "test-host-id",
          path: "/var/www",
          destination: "backup-dest",
          schedule: "0 0 * * *",
          excludes: ["*.log", "tmp/"],
          incremental: false,
          retention: 4,
          encrypted: false,
          hooks: [],
        });
        expect(mockPush).toHaveBeenCalledWith("/hosts/test-host-id");
      });
    });

    it("displays error message on submission failure", async () => {
      const user = userEvent.setup();
      (createBackup as any).mockRejectedValue(new Error("Network error"));

      render(<CreateBackupForm hostId="test-host-id" />);

      const destInput = screen.getByLabelText("Destination Name");
      await user.type(destInput, "backup");

      const pathInput = screen.getByLabelText("Source Path");
      await user.type(pathInput, "/var/www");

      await user.click(
        screen.getByRole("button", { name: "Create Backup Task" }),
      );

      await waitFor(() => {
        expect(screen.getByText("Network error")).toBeInTheDocument();
      });

      expect(mockPush).not.toHaveBeenCalled();
    });

    it("includes incremental flag when switch is enabled", async () => {
      const user = userEvent.setup();
      (createBackup as any).mockResolvedValue({});

      render(<CreateBackupForm hostId="test-host-id" />);

      const destInput = screen.getByLabelText("Destination Name");
      await user.type(destInput, "backup");

      const pathInput = screen.getByLabelText("Source Path");
      await user.type(pathInput, "/var/www");

      await user.click(screen.getByText("Advanced Settings"));
      await user.click(
        screen.getByRole("switch", { name: /incremental backup/i }),
      );

      await user.click(
        screen.getByRole("button", { name: "Create Backup Task" }),
      );

      await waitFor(() => {
        expect(createBackup).toHaveBeenCalledWith(
          expect.objectContaining({
            incremental: true,
            retention: 4,
            encrypted: false,
          }),
        );
      });
    });

    it("includes encrypted flag when switch is enabled", async () => {
      const user = userEvent.setup();
      (createBackup as any).mockResolvedValue({});

      render(<CreateBackupForm hostId="test-host-id" />);

      const destInput = screen.getByLabelText("Destination Name");
      await user.type(destInput, "backup");

      const pathInput = screen.getByLabelText("Source Path");
      await user.type(pathInput, "/var/www");

      await user.click(screen.getByText("Advanced Settings"));
      await user.click(screen.getByRole("switch", { name: /encrypt backup/i }));

      await user.click(
        screen.getByRole("button", { name: "Create Backup Task" }),
      );

      await waitFor(() => {
        expect(createBackup).toHaveBeenCalledWith(
          expect.objectContaining({
            encrypted: true,
          }),
        );
      });
    });
  });
});
