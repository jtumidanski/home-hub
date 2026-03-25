import { createContext, useContext, useId, type ReactNode } from "react";
import { cn } from "@/lib/utils";

interface RadioGroupContextValue {
  value: string;
  onValueChange: (value: string) => void;
  name: string;
}

const RadioGroupContext = createContext<RadioGroupContextValue | undefined>(undefined);

interface RadioGroupProps {
  value: string;
  onValueChange: (value: string) => void;
  className?: string;
  children: ReactNode;
}

export function RadioGroup({ value, onValueChange, className, children }: RadioGroupProps) {
  const name = useId();
  return (
    <RadioGroupContext.Provider value={{ value, onValueChange, name }}>
      <div role="radiogroup" className={cn("flex gap-4", className)}>
        {children}
      </div>
    </RadioGroupContext.Provider>
  );
}

interface RadioGroupItemProps {
  value: string;
  id?: string;
  className?: string;
  children: ReactNode;
}

export function RadioGroupItem({ value, id, className, children }: RadioGroupItemProps) {
  const context = useContext(RadioGroupContext);
  if (!context) {
    throw new Error("RadioGroupItem must be used within RadioGroup");
  }
  const itemId = id ?? `${context.name}-${value}`;

  return (
    <label htmlFor={itemId} className={cn("flex cursor-pointer items-center gap-2", className)}>
      <input
        type="radio"
        id={itemId}
        name={context.name}
        value={value}
        checked={context.value === value}
        onChange={() => context.onValueChange(value)}
        className="h-4 w-4 border-border text-primary focus:ring-ring"
      />
      <span className="text-sm">{children}</span>
    </label>
  );
}
