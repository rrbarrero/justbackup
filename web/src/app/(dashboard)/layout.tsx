import { AppSidebar } from "@/shared/components/sidebar";
import { SidebarProvider, SidebarTrigger } from "@/shared/ui/sidebar";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <SidebarProvider>
      <AppSidebar />
      <main className="flex-1 overflow-y-auto p-8">
        <SidebarTrigger />
        {children}
      </main>
    </SidebarProvider>
  );
}
