import * as React from "react";
import { cn } from "@/lib/cn";

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "secondary" | "ghost" | "outline" | "destructive";
  size?: "sm" | "md" | "icon";
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "md", ...props }, ref) => (
    <button
      ref={ref}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-md border text-sm font-semibold transition-[background,box-shadow,transform] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
        "active:translate-x-[1px] active:translate-y-[1px]",
        variant === "default" &&
          "border-primary/70 bg-primary text-primary-foreground shadow-[2px_3px_0_hsl(214_45%_30%/0.16)] hover:bg-primary/90",
        variant === "secondary" &&
          "border-border bg-secondary text-secondary-foreground shadow-[2px_3px_0_hsl(218_32%_30%/0.1)] hover:bg-accent",
        variant === "ghost" && "border-transparent bg-transparent hover:border-border hover:bg-accent/70",
        variant === "outline" &&
          "border-input bg-card/90 text-foreground shadow-[2px_3px_0_hsl(218_32%_30%/0.08)] hover:bg-accent/70",
        variant === "destructive" &&
          "border-destructive/70 bg-destructive text-destructive-foreground shadow-[2px_3px_0_hsl(1_45%_30%/0.16)] hover:bg-destructive/90",
        size === "sm" && "h-8 px-3",
        size === "md" && "h-9 px-4",
        size === "icon" && "h-9 w-9",
        className,
      )}
      {...props}
    />
  ),
);
Button.displayName = "Button";
