import { AddHostForm } from "@/host/components/add-host-form";

export default function AddHostPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">New Host</h1>
      </div>
      <AddHostForm />
    </div>
  );
}
