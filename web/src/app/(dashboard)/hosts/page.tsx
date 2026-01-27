import Link from "next/link";
import { Button } from "@/shared/ui/button";
import { HostsTable } from "@/host/components/hosts-table";

export default function HostsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Hosts</h1>
        <Link href="/hosts/new">
          <Button>Add Host</Button>
        </Link>
      </div>
      <HostsTable />
    </div>
  );
}
