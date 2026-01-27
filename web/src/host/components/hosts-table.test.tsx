import { render, screen, waitFor } from "@testing-library/react";
import { HostsTable } from "./hosts-table";
import { getHosts } from "@/host/infrastructure/host-api";
import { vi } from "vitest";

// Mock the API module
vi.mock("@/host/infrastructure/host-api", () => ({
  getHosts: vi.fn(),
}));

// Mock useRouter
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

// Mock the WebSocket hook
const mockUseWebSocket = vi.fn();
vi.mock("@/shared/hooks/use-websocket", () => ({
  useWebSocket: (...args: any[]) => mockUseWebSocket(...args),
}));

describe("HostsTable", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseWebSocket.mockReturnValue(undefined);
  });

  it("renders loading state initially", () => {
    // Mock getHosts to return a promise that doesn't resolve immediately
    (getHosts as any).mockReturnValue(new Promise(() => {}));

    render(<HostsTable />);

    expect(screen.getByText("Loading hosts...")).toBeInTheDocument();
  });

  it("renders empty state when no hosts are returned", async () => {
    (getHosts as any).mockResolvedValue([]);

    render(<HostsTable />);

    await waitFor(() => {
      expect(
        screen.getByText("No hosts found. Add your first host to get started."),
      ).toBeInTheDocument();
    });
  });

  it("renders table with hosts when data is returned", async () => {
    const mockHosts = [
      {
        id: "1",
        name: "Host 1",
        hostname: "192.168.1.1",
        user: "root",
        port: 22,
      },
      {
        id: "2",
        name: "Host 2",
        hostname: "192.168.1.2",
        user: "admin",
        port: 2222,
      },
    ];
    (getHosts as any).mockResolvedValue(mockHosts);

    render(<HostsTable />);

    await waitFor(() => {
      expect(screen.getByText("Host 1")).toBeInTheDocument();
      expect(screen.getByText("192.168.1.1")).toBeInTheDocument();
      expect(screen.getByText("Host 2")).toBeInTheDocument();
      expect(screen.getByText("192.168.1.2")).toBeInTheDocument();
    });
  });

  it("renders error state when API fails", async () => {
    (getHosts as any).mockRejectedValue(new Error("API Error"));

    render(<HostsTable />);

    await waitFor(() => {
      expect(screen.getByText("Error: API Error")).toBeInTheDocument();
    });
  });
});
