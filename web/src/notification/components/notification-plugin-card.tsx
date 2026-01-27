import React from "react";
import Image from "next/image";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Switch } from "@/shared/ui/switch";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card";
import { Collapsible, CollapsibleContent } from "@/shared/ui/collapsible";
import { cn } from "@/shared/lib/utils";

interface NotificationPluginCardProps {
  title: string;
  description: string;
  logoSrc: string;
  isEnabled: boolean;
  onEnableChange: (enabled: boolean) => void;
  children: React.ReactNode;
  className?: string;
}

export function NotificationPluginCard({
  title,
  description,
  logoSrc,
  isEnabled,
  onEnableChange,
  children,
  className,
}: NotificationPluginCardProps) {
  const [isExpanded, setIsExpanded] = React.useState(isEnabled);

  // Auto-expand when enabled
  React.useEffect(() => {
    if (isEnabled) {
      setIsExpanded(true);
    }
  }, [isEnabled]);

  return (
    <Card className={cn("overflow-hidden", className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div
          className="flex items-center gap-4 cursor-pointer flex-1"
          onClick={() => setIsExpanded(!isExpanded)}
        >
          <div className="relative h-10 w-10 overflow-hidden rounded-md border bg-muted/50 p-1">
            <Image
              src={logoSrc}
              alt={`${title} logo`}
              fill
              className="object-contain p-1"
            />
          </div>
          <div className="space-y-1">
            <CardTitle className="text-base flex items-center gap-2">
              {title}
              {isExpanded ? (
                <ChevronUp className="h-4 w-4 text-muted-foreground" />
              ) : (
                <ChevronDown className="h-4 w-4 text-muted-foreground" />
              )}
            </CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Switch
            checked={isEnabled}
            onCheckedChange={(checked) => {
              onEnableChange(checked);
              if (checked) setIsExpanded(true);
            }}
          />
        </div>
      </CardHeader>
      <Collapsible open={isExpanded}>
        <CollapsibleContent>
          <CardContent className="pt-4 border-t mt-4 bg-muted/20">
            {children}
          </CardContent>
        </CollapsibleContent>
      </Collapsible>
    </Card>
  );
}
