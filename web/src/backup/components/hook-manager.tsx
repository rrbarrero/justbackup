"use client";

import {
  useFieldArray,
  Control,
  UseFormRegister,
  useWatch,
  UseFormSetValue,
} from "react-hook-form";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Switch } from "@/shared/ui/switch";
import {
  Plus,
  Trash2,
  Settings2,
  ChevronDown,
  ChevronUp,
  AlertCircle,
  ChevronRight,
  Database,
  Eye,
  EyeOff,
} from "lucide-react";
import { useState, useEffect } from "react";
import { cn } from "@/shared/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";

interface HookData {
  name: string;
  phase: "pre" | "post";
  enabled: boolean;
  params: Record<string, string>;
}

const HOOK_TEMPLATES = [
  {
    label: "PostgreSQL Dump",
    name: "postgres_dump",
    params: {
      HOST: "localhost",
      PORT: "5432",
      USER: "postgres",
      PASSWORD: "",
      DB_NAME: "mydb",
    },
  },
  {
    label: "MySQL Dump",
    name: "mysql_dump",
    params: {
      HOST: "localhost",
      PORT: "3306",
      USER: "root",
      PASSWORD: "",
      DB_NAME: "mydb",
    },
  },
  {
    label: "MongoDB Dump",
    name: "mongodb_dump",
    params: {
      HOST: "localhost",
      PORT: "27017",
      USER: "",
      PASSWORD: "",
      DB_NAME: "mydb",
      AUTH_DB: "admin",
    },
  },
  {
    label: "Custom Script",
    name: "custom_script",
    params: {},
  },
];

interface HookManagerProps {
  control: Control<any>;
  register: UseFormRegister<any>;
  setValue: UseFormSetValue<any>;
}

export function HookManager({ control, register, setValue }: HookManagerProps) {
  const { fields, append, remove } = useFieldArray({
    control,
    name: "hooks",
  });

  const watchedHooks = useWatch({
    control,
    name: "hooks",
  });

  const [expandedIndex, setExpandedIndex] = useState<number | null>(null);

  const addHookFromTemplate = (template: (typeof HOOK_TEMPLATES)[0]) => {
    append({
      name: template.name,
      phase: "pre",
      enabled: true,
      params: template.params,
    });
    setExpandedIndex(fields.length);
  };

  const toggleExpand = (index: number) => {
    setExpandedIndex(expandedIndex === index ? null : index);
  };

  return (
    <div className="space-y-4 border-t pt-4">
      <div className="flex flex-col space-y-1">
        <div className="flex items-center justify-between">
          <Label className="text-base uppercase text-xs font-bold text-muted-foreground tracking-wider">
            Backup Hooks
          </Label>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="h-7 text-xs"
              >
                <Plus className="h-3 w-3 mr-1" /> Add Hook
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel className="text-[10px] uppercase font-bold text-muted-foreground">
                Select Template
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              {HOOK_TEMPLATES.map((template) => (
                <DropdownMenuItem
                  key={template.name}
                  onClick={() => addHookFromTemplate(template)}
                  className="flex items-center justify-between cursor-pointer"
                >
                  <div className="flex items-center gap-2">
                    <Database className="h-3.5 w-3.5 text-primary" />
                    <span className="text-sm">{template.label}</span>
                  </div>
                  <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <p className="text-[11px] text-muted-foreground leading-tight">
          Run scripts (database dumps, notifications, etc.) from{" "}
          <code>/app/plugins</code> on the worker automatically.
        </p>
      </div>

      {fields.length === 0 && (
        <div className="text-center py-4 border-2 border-dashed rounded-md bg-muted/20">
          <p className="text-xs text-muted-foreground">No hooks configured</p>
        </div>
      )}

      <div className="space-y-2">
        {fields.map((field, index) => (
          <HookItem
            key={field.id}
            index={index}
            hookData={(watchedHooks?.[index] || {}) as HookData}
            register={register}
            control={control}
            setValue={setValue}
            remove={remove}
            isExpanded={expandedIndex === index}
            toggleExpand={() => toggleExpand(index)}
          />
        ))}
      </div>
    </div>
  );
}

interface HookItemProps {
  index: number;
  hookData: HookData;
  register: UseFormRegister<any>;
  control: Control<any>;
  setValue: UseFormSetValue<any>;
  remove: (index: number) => void;
  isExpanded: boolean;
  toggleExpand: () => void;
}

function HookItem({
  index,
  hookData,
  register,
  control,
  setValue,
  remove,
  isExpanded,
  toggleExpand,
}: HookItemProps) {
  const [localJson, setLocalJson] = useState("");
  const [jsonError, setJsonError] = useState<string | null>(null);
  const [showParams, setShowParams] = useState(false);

  // Sync local text when hookData changes (but only if not focused to avoid cycle,
  // though here we simplified to a basic sync)
  useEffect(() => {
    if (hookData.params && Object.keys(hookData.params).length > 0) {
      setLocalJson(JSON.stringify(hookData.params, null, 2));
    } else {
      setLocalJson("");
    }
    setJsonError(null);
  }, [hookData.params]);

  const handleJsonChange = (val: string) => {
    setLocalJson(val);
    if (!val.trim()) {
      setJsonError(null);
      updateParams({});
      return;
    }

    try {
      const parsed = JSON.parse(val);
      if (
        typeof parsed !== "object" ||
        parsed === null ||
        Array.isArray(parsed)
      ) {
        throw new Error("Parameters must be a JSON object");
      }

      const stringifiedParams: Record<string, string> = {};
      for (const k in parsed) {
        stringifiedParams[k] = String(parsed[k]);
      }

      setJsonError(null);
      updateParams(stringifiedParams);
    } catch (err: any) {
      setJsonError(err.message);
    }
  };

  const updateParams = (params: Record<string, string>) => {
    setValue(`hooks.${index}.params` as any, params, {
      shouldValidate: true,
      shouldDirty: true,
      shouldTouch: true,
    });
  };

  return (
    <div className="border rounded-md overflow-hidden bg-card shadow-sm">
      <div
        className="flex items-center gap-2 px-3 py-2 cursor-pointer hover:bg-muted"
        onClick={toggleExpand}
      >
        <div className="flex items-center gap-2 flex-1">
          <Settings2 className="h-4 w-4 text-muted-foreground" />
          <span className="font-medium text-sm">
            {hookData.name || "New Hook"}
          </span>
          <span className="text-[10px] bg-muted px-1.5 py-0.5 rounded text-muted-foreground uppercase font-bold">
            {hookData.phase}
          </span>
          {jsonError && (
            <AlertCircle className="h-4 w-4 text-red-500 animate-pulse" />
          )}
        </div>
        <div className="flex items-center gap-1">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={(e) => {
              e.stopPropagation();
              remove(index);
            }}
            className="h-7 w-7 text-red-500 hover:text-red-600 hover:bg-red-50"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
          {isExpanded ? (
            <ChevronUp className="h-4 w-4 text-muted-foreground" />
          ) : (
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          )}
        </div>
      </div>

      {isExpanded && (
        <div className="p-4 border-t bg-muted/10 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label className="text-[10px] uppercase font-bold text-muted-foreground">
                Script Name
              </Label>
              <Input
                {...register(`hooks.${index}.name` as const)}
                placeholder="e.g. postgres_dump"
                className="h-8 text-sm"
              />
            </div>
            <div className="space-y-2">
              <Label className="text-[10px] uppercase font-bold text-gray-500">
                Phase
              </Label>
              <select
                {...register(`hooks.${index}.phase` as const)}
                className="w-full h-8 text-sm px-2 border rounded-md bg-muted border-input focus:outline-none focus:ring-2 focus:ring-primary/20"
              >
                <option value="pre">Pre-backup</option>
                <option value="post">Post-backup</option>
              </select>
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id={`hook-enabled-${index}`}
              checked={hookData.enabled}
              onCheckedChange={(checked) => {
                const name = `hooks.${index}.enabled` as const;
                const fields = (control as any)._fields;
                if (fields[name]) {
                  fields[name]._f.onChange({
                    target: { checked, name },
                  });
                }
              }}
            />
            <Label htmlFor={`hook-enabled-${index}`} className="text-sm">
              Enabled
            </Label>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label
                className={cn(
                  "text-[10px] uppercase font-bold text-muted-foreground",
                  jsonError && "text-destructive",
                )}
              >
                Parameters (JSON)
              </Label>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-6 px-2 text-[10px] uppercase font-bold text-muted-foreground hover:text-foreground"
                onClick={() => setShowParams(!showParams)}
              >
                {showParams ? (
                  <>
                    <EyeOff className="h-3 w-3 mr-1" /> Hide
                  </>
                ) : (
                  <>
                    <Eye className="h-3 w-3 mr-1" /> Show
                  </>
                )}
              </Button>
            </div>
            {showParams ? (
              <>
                <textarea
                  className={cn(
                    "w-full min-h-[100px] text-xs font-mono p-2 border rounded-md bg-muted border-input focus:outline-none focus:ring-2 focus:ring-primary/20",
                    jsonError && "border-destructive focus:ring-destructive/20",
                  )}
                  placeholder='{
  "DB_NAME": "myapp_db",
  "DB_USER": "root",
  "DB_PASSWORD": "password123"
}'
                  value={localJson}
                  onChange={(e) => handleJsonChange(e.target.value)}
                />
                {jsonError ? (
                  <p className="text-[10px] text-destructive font-medium">
                    Invalid JSON: {jsonError}
                  </p>
                ) : (
                  <p className="text-[10px] text-muted-foreground italic">
                    Injects <code>HOOK_PARAM_KEY=VALUE</code> into the script.
                  </p>
                )}
              </>
            ) : (
              <div
                className="w-full min-h-[100px] flex flex-col items-center justify-center border border-dashed rounded-md bg-muted/30 cursor-pointer hover:bg-muted/50 transition-colors"
                onClick={() => setShowParams(true)}
              >
                <Eye className="h-5 w-5 text-muted-foreground mb-1" />
                <p className="text-[10px] text-muted-foreground uppercase font-bold">
                  Parameters are hidden
                </p>
                <p className="text-[9px] text-muted-foreground/60 italic">
                  Click to reveal sensitive information
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
