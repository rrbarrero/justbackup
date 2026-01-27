"use client";

import { useEffect, useState } from "react";
import {
  getSSHKey,
  generateToken,
  getTokenStatus,
  revokeToken,
  TokenResponse,
} from "@/settings/infrastructure/settings-api";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { Copy, Check, AlertTriangle, RefreshCw, Trash2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/shared/ui/alert-dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";
import { NotificationSettingsForm } from "@/notification/components/notification-settings-form";

export default function SettingsPage() {
  const [publicKey, setPublicKey] = useState<string>("");
  const [tokenStatus, setTokenStatus] = useState<TokenResponse | null>(null);
  const [newToken, setNewToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [tokenCopied, setTokenCopied] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [sshKeyData, tokenData] = await Promise.all([
          getSSHKey(),
          getTokenStatus(),
        ]);
        setPublicKey(sshKeyData.publicKey);
        setTokenStatus(tokenData);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load settings",
        );
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(publicKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
  };

  const handleGenerateToken = async () => {
    try {
      const data = await generateToken();
      setTokenStatus(data);
      setNewToken(data.token || null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to generate token");
    }
  };

  const handleRevokeToken = async () => {
    try {
      await revokeToken();
      setTokenStatus({ exists: false });
      setNewToken(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to revoke token");
    }
  };

  const handleCopyToken = async () => {
    if (!newToken) return;
    try {
      await navigator.clipboard.writeText(newToken);
      setTokenCopied(true);
      setTimeout(() => setTokenCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy token: ", err);
    }
  };

  if (isLoading) {
    return <div className="p-4 text-center">Loading settings...</div>;
  }

  if (error) {
    return <div className="p-4 text-red-500">Error: {error}</div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Settings</h1>

      <Tabs defaultValue="general" className="w-full">
        <TabsList className="grid w-full grid-cols-2 max-w-[400px]">
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="notifications">Notifications</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="space-y-6 mt-6">
          <Card>
            <CardHeader>
              <CardTitle>SSH Configuration</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <div>
                  <h3 className="text-sm font-medium mb-1">
                    1. Secured Authorization Line (Recommended)
                  </h3>
                  <p className="text-xs text-muted-foreground mb-3">
                    This line combines your public key with strict security
                    rules. It ensures JustBackup can <strong>only</strong>{" "}
                    execute backups and size checks, nothing else.
                  </p>
                  <div className="relative">
                    <pre className="bg-zinc-950 text-zinc-50 dark:bg-zinc-900 p-4 rounded-md overflow-x-auto text-xs font-mono whitespace-pre-wrap break-all pr-12 border border-blue-500/20 shadow-sm">
                      {`command="case \\"$SSH_ORIGINAL_COMMAND\\" in rsync\\ --server*|du\\ -sk*) $SSH_ORIGINAL_COMMAND ;; *) echo \\"Access Denied by JustBackup\\"; exit 1 ;; esac",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty ${publicKey}`}
                    </pre>
                    <Button
                      size="icon"
                      variant="ghost"
                      className="absolute top-2 right-2 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800"
                      onClick={() => {
                        const fullKey = `command="case \\"$SSH_ORIGINAL_COMMAND\\" in rsync\\ --server*|du\\ -sk*) $SSH_ORIGINAL_COMMAND ;; *) echo \\"Access Denied by JustBackup\\"; exit 1 ;; esac",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty ${publicKey}`;
                        navigator.clipboard.writeText(fullKey);
                        setCopied(true);
                        setTimeout(() => setCopied(false), 2000);
                      }}
                    >
                      {copied ? (
                        <Check className="h-4 w-4 text-green-500" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>

                <div className="bg-muted/50 rounded-lg p-4 border text-sm space-y-3">
                  <h4 className="font-semibold flex items-center gap-2">
                    <Check className="h-4 w-4 text-primary" />
                    Installation Instructions
                  </h4>
                  <ol className="list-decimal list-inside space-y-1 text-muted-foreground ml-1">
                    <li>Log in to your remote server via SSH.</li>
                    <li>
                      Open the <code>~/.ssh/authorized_keys</code> file.
                    </li>
                    <li>
                      Paste the <strong>Secured Authorization Line</strong>{" "}
                      (from above) on a new line.
                    </li>
                    <li>Save the file. You&apos;re done!</li>
                  </ol>
                </div>
              </div>

              <div className="pt-4 border-t">
                <details className="group">
                  <summary className="flex items-center gap-2 text-xs font-medium text-muted-foreground cursor-pointer hover:text-foreground transition-colors select-none">
                    <span>Advanced: View Raw Public Key</span>
                  </summary>
                  <div className="mt-3 relative">
                    <p className="text-xs text-muted-foreground mb-2">
                      This is the raw SSH key without any security wrappers. Use
                      this only if you know what you are doing or need it for
                      manual configuration debugging.
                    </p>
                    <pre className="bg-muted p-3 rounded-md overflow-x-auto text-[10px] font-mono break-all pr-10 text-muted-foreground group-open:text-foreground transition-colors">
                      {publicKey}
                    </pre>
                    <Button
                      size="icon"
                      variant="ghost"
                      className="absolute top-8 right-2 h-6 w-6"
                      onClick={handleCopy}
                    >
                      {copied ? (
                        <Check className="h-3 w-3 text-green-500" />
                      ) : (
                        <Copy className="h-3 w-3" />
                      )}
                    </Button>
                  </div>
                </details>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>CLI Access Token</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-muted-foreground">
                Generate an API token to authenticate the JustBackup CLI tool
                with this server.
              </p>
              {newToken && (
                <Alert className="border-green-500/20 bg-green-500/10">
                  <Check className="h-4 w-4 text-green-500" />
                  <AlertTitle className="text-green-600 dark:text-green-500">
                    Token Generated Successfully
                  </AlertTitle>
                  <AlertDescription>
                    <p className="mb-2 text-green-700 dark:text-green-400">
                      Please copy your token now. For security reasons, it will
                      not be shown again.
                    </p>
                    <div className="relative">
                      <pre className="bg-muted p-3 rounded border border-border overflow-x-auto text-sm font-mono whitespace-pre-wrap break-all pr-12 text-foreground">
                        {newToken}
                      </pre>
                      <Button
                        size="icon"
                        variant="ghost"
                        className="absolute top-1 right-1 h-8 w-8 hover:bg-green-500/10"
                        onClick={handleCopyToken}
                      >
                        {tokenCopied ? (
                          <Check className="h-4 w-4 text-green-500" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </AlertDescription>
                </Alert>
              )}

              {!tokenStatus?.exists ? (
                <div className="flex flex-col items-start gap-4">
                  <p className="text-sm">
                    You don&apos;t have an active token. Generate one to start
                    using the CLI.
                  </p>
                  <Button onClick={handleGenerateToken}>Generate Token</Button>
                </div>
              ) : (
                <div className="space-y-4">
                  <div className="flex items-center justify-between bg-muted/50 p-4 rounded-lg border">
                    <div>
                      <h4 className="font-medium">Active Token</h4>
                      <p className="text-sm text-muted-foreground">
                        Created on{" "}
                        {tokenStatus.created_at
                          ? new Date(tokenStatus.created_at).toLocaleString()
                          : "Unknown"}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button
                            variant="outline"
                            className="text-orange-500 dark:text-orange-400 border-orange-500/20 hover:bg-orange-500/10"
                          >
                            <RefreshCw className="mr-2 h-4 w-4" />
                            Regenerate
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>
                              Regenerate Token?
                            </AlertDialogTitle>
                            <AlertDialogDescription>
                              This will invalidate your current token
                              immediately. Any CLI tools using the old token
                              will stop working until updated.
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction onClick={handleGenerateToken}>
                              Regenerate
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>

                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button variant="destructive">
                            <Trash2 className="mr-2 h-4 w-4" />
                            Revoke
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>Revoke Token?</AlertDialogTitle>
                            <AlertDialogDescription>
                              Are you sure you want to revoke this token? The
                              CLI will no longer be able to access the server.
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction
                              onClick={handleRevokeToken}
                              className="bg-red-600 hover:bg-red-700"
                            >
                              Revoke
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </div>
                  </div>
                  {!newToken && (
                    <Alert
                      variant="default"
                      className="bg-blue-500/10 border-blue-500/20"
                    >
                      <AlertTriangle className="h-4 w-4 text-blue-500" />
                      <AlertTitle className="text-blue-600 dark:text-blue-500">
                        Note
                      </AlertTitle>
                      <AlertDescription className="text-blue-600 dark:text-blue-400">
                        Your token is hidden for security. If you lost it, you
                        must regenerate a new one.
                      </AlertDescription>
                    </Alert>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="notifications" className="mt-6">
          <NotificationSettingsForm />
        </TabsContent>
      </Tabs>
    </div>
  );
}
