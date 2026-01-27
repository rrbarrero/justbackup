"use client";

import { Button } from "@/shared/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import {
  Terminal,
  Download,
  Info,
  Search,
  RotateCcw,
  FolderTree,
  Play,
  List,
  Settings2,
  Key,
  ShieldCheck,
  Globe,
  Monitor,
} from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";

export default function CliPage() {
  return (
    <div className="space-y-8 pb-10">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          Command Line Interface
        </h1>
        <p className="text-muted-foreground mt-2">
          Automate and manage your backup infrastructure directly from your
          terminal.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="flex flex-col">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Download className="h-5 w-5 text-primary" />
              Installation
            </CardTitle>
            <CardDescription>
              Quickly install the JustBackup binary on your local machine.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 flex-1">
            <div className="rounded-md bg-muted p-4 font-mono text-xs overflow-x-auto">
              <pre>
                <code>
                  {`# Download the binary
curl -L ${typeof window !== "undefined" ? window.location.origin : ""}/downloads/justbackup -o justbackup

# Make it executable
chmod +x justbackup

# Move to your path (optional)
sudo mv justbackup /usr/local/bin/`}
                </code>
              </pre>
            </div>
            <Button asChild className="w-full">
              <a href="/downloads/justbackup" download="justbackup">
                <Download className="mr-2 h-4 w-4" />
                Download Binary (Linux/macOS)
              </a>
            </Button>
            <Alert className="pl-12 border-blue-500/20 bg-blue-50/50 dark:bg-blue-950/10">
              <Info className="h-4 w-4 !text-blue-600 dark:!text-blue-400" />
              <AlertTitle className="text-blue-700 dark:text-blue-400">
                Secure Compilation
              </AlertTitle>
              <AlertDescription className="text-blue-600/80 dark:text-blue-400/80 text-xs">
                This binary is compiled from source during deployment, ensuring
                1:1 parity with your server's version.
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>

        <Card className="flex flex-col">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings2 className="h-5 w-5 text-primary" />
              Initial Setup
            </CardTitle>
            <CardDescription>
              Connect your CLI to this JustBackup instance.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 flex-1">
            <p className="text-sm text-muted-foreground">
              Run the configuration command and follow the prompts. You'll need
              your API token.
            </p>
            <div className="rounded-md bg-zinc-900 p-4 font-mono text-sm text-zinc-100">
              <code>justbackup config</code>
            </div>
            <div className="space-y-3 pt-2">
              <div className="flex items-start gap-2 text-sm">
                <div className="bg-primary/10 p-1 rounded mt-0.5">
                  <Globe className="h-3 w-3 text-primary" />
                </div>
                <div>
                  <span className="font-semibold">Backend URL:</span>
                  <p className="text-muted-foreground text-xs whitespace-nowrap overflow-hidden text-ellipsis">
                    {typeof window !== "undefined"
                      ? window.location.origin
                      : "https://your-instance.com"}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-2 text-sm">
                <div className="bg-primary/10 p-1 rounded mt-0.5">
                  <Key className="h-3 w-3 text-primary" />
                </div>
                <div>
                  <span className="font-semibold">API Token:</span>
                  <p className="text-muted-foreground text-xs">
                    Found in{" "}
                    <a
                      href="/settings"
                      className="text-primary hover:underline"
                    >
                      Settings
                    </a>
                    .
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Terminal className="h-5 w-5 text-primary" />
            Command Reference
          </CardTitle>
          <CardDescription>
            Detailed documentation for all available commands.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="management" className="w-full">
            <TabsList className="grid w-full grid-cols-3 mb-6">
              <TabsTrigger value="management">Infrastructure</TabsTrigger>
              <TabsTrigger value="exploration">Data & Files</TabsTrigger>
              <TabsTrigger value="actions">Restore & Jobs</TabsTrigger>
            </TabsList>

            <TabsContent value="management" className="space-y-4">
              <div className="grid gap-4">
                <CommandItem
                  command="justbackup hosts"
                  description="List all registered hosts, showing their status, hostname, and connection details."
                  icon={<List className="h-4 w-4" />}
                />
                <CommandItem
                  command="justbackup bootstrap"
                  args="--host <ip> --name <alias> [--user <user>] [--port <port>]"
                  description="Automatic SSH key installation and host registration. It proactively secures the installed key in authorized_keys, restricting it to only execute rsync and du commands."
                  icon={<ShieldCheck className="h-4 w-4" />}
                  example="justbackup bootstrap --host 192.168.1.50 --name web-server --user root"
                />
                <CommandItem
                  command="justbackup add-backup"
                  args="--host-id <id> --path <path> --dest <name> [options]"
                  description="Create a new scheduled backup task for a specific host. You can define cron schedules and exclude patterns."
                  icon={<Settings2 className="h-4 w-4" />}
                  example='justbackup add-backup --host-id h123 --path /home/user/docs --dest documents --schedule "0 2 * * *"'
                />
              </div>
            </TabsContent>

            <TabsContent value="exploration" className="space-y-4">
              <div className="grid gap-4">
                <CommandItem
                  command="justbackup backups"
                  args="[host-id]"
                  description="List recent backups. Provide a host ID to filter the results."
                  icon={<FolderTree className="h-4 w-4" />}
                />
                <CommandItem
                  command="justbackup files"
                  args="<backup-id> [--path <subpath>]"
                  description="Browse files inside a specific backup. You can navigate through directories using the --path flag."
                  icon={<List className="h-4 w-4" />}
                  example="justbackup files b123... --path /var/www"
                />
                <CommandItem
                  command="justbackup search"
                  args="<pattern>"
                  description="Global search across all backups for files matching the given pattern."
                  icon={<Search className="h-4 w-4" />}
                  example="justbackup search config.json"
                />
              </div>
            </TabsContent>

            <TabsContent value="actions" className="space-y-4">
              <div className="grid gap-4">
                <CommandItem
                  command="justbackup run"
                  args="<backup-id>"
                  description="Immediately trigger an asynchronous backup task for the specified configuration."
                  icon={<Play className="h-4 w-4" />}
                />
                <div className="border rounded-lg p-4 space-y-4 bg-muted/30">
                  <div className="flex items-center gap-2 font-semibold text-primary">
                    <RotateCcw className="h-5 w-5" />
                    Restore Commands
                  </div>
                  <p className="text-sm text-muted-foreground">
                    The restore command is powerful and supports two main modes
                    of operation:
                  </p>

                  <div className="grid gap-4 sm:grid-cols-2">
                    <div className="space-y-2">
                      <h4 className="text-xs font-bold uppercase tracking-wider flex items-center gap-1">
                        <Monitor className="h-3 w-3" /> Local Restore
                      </h4>
                      <p className="text-xs text-muted-foreground">
                        Downloads files directly to your current machine using a
                        secure one-time tunnel.
                      </p>
                      <code className="text-xs block bg-zinc-900 text-zinc-100 p-2 rounded">
                        justbackup restore &lt;id&gt; --local --path /etc --dest
                        ./restored
                      </code>
                    </div>
                    <div className="space-y-2">
                      <h4 className="text-xs font-bold uppercase tracking-wider flex items-center gap-1">
                        <Globe className="h-3 w-3" /> Remote Restore
                      </h4>
                      <p className="text-xs text-muted-foreground">
                        Tells the worker to rsync files directly to another host
                        (or the same one at a different path).
                      </p>
                      <code className="text-xs block bg-zinc-900 text-zinc-100 p-2 rounded">
                        justbackup restore &lt;id&gt; --remote --path /app
                        --to-path /app_recovered
                      </code>
                    </div>
                  </div>
                </div>
                <CommandItem
                  command="justbackup decrypt"
                  args="--file <path> --out <path> --id <backup-id> --key <master-key>"
                  description="Offline decryption of a backup file without requiring backend access. Essential for disaster recovery when the server is unavailable."
                  icon={<ShieldCheck className="h-4 w-4" />}
                  example="justbackup decrypt --file backup.tar.gz.enc --out restore.tar.gz --id uuid-123 --key my-secret-key"
                />
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}

function CommandItem({
  command,
  args,
  description,
  icon,
  example,
}: {
  command: string;
  args?: string;
  description: string;
  icon: React.ReactNode;
  example?: string;
}) {
  return (
    <div className="flex flex-col gap-2 p-4 border rounded-lg hover:border-primary/50 transition-colors">
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="bg-primary/10 p-2 rounded text-primary">{icon}</div>
          <div className="flex flex-wrap items-center gap-2">
            <code className="font-mono font-bold text-sm bg-muted px-2 py-0.5 rounded uppercase">
              {command}
            </code>
            {args && (
              <code className="font-mono text-xs text-muted-foreground">
                {args}
              </code>
            )}
          </div>
        </div>
      </div>
      <p className="text-sm text-muted-foreground">{description}</p>
      {example && (
        <div className="mt-1">
          <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
            Example:
          </span>
          <code className="block mt-1 text-xs font-mono bg-zinc-100 dark:bg-zinc-800 p-2 rounded border border-dashed text-primary/80">
            {example}
          </code>
        </div>
      )}
    </div>
  );
}
