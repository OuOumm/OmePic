import * as React from "react";

import { cn } from "@/lib/utils";

type CardVariant = "default" | "strong" | "subtle";

const cardVariants: Record<CardVariant, string> = {
  default: "border border-border bg-card text-card-foreground shadow-sm",
  strong: "border border-border bg-card text-card-foreground shadow-sm",
  subtle: "border border-border bg-muted/30 text-card-foreground shadow-sm"
};

type CardProps = React.HTMLAttributes<HTMLDivElement> & {
  variant?: CardVariant;
};

function Card({ className, variant = "default", ...props }: CardProps) {
  return <div className={cn("rounded-lg", cardVariants[variant], className)} {...props} />;
}

function CardHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex flex-col space-y-1.5 p-6", className)} {...props} />;
}

function CardTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className={cn("text-2xl font-semibold leading-none tracking-tight", className)}
      {...props}
    />
  );
}

function CardDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn("text-sm text-muted-foreground", className)} {...props} />;
}

function CardContent({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("p-6 pt-0", className)} {...props} />;
}

function CardFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex items-center p-6 pt-0", className)} {...props} />;
}

export { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle };
