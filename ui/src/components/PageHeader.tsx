export interface PageHeaderProps {
    title: string;
    description?: string;
    children?: React.ReactNode;
}

export function PageHeader({ title, description, children }: PageHeaderProps) {
    return (
        <div className="flex items-center justify-between p-6 border-b border-border bg-background/50 backdrop-blur-sm sticky top-0 z-20 shrink-0">
            <div className="flex flex-col gap-1">
                <h1 className="text-xl font-bold font-mono tracking-tight text-foreground/90">
                    {title}
                </h1>
                {description && (
                    <p className="text-sm text-muted-foreground font-medium">
                        {description}
                    </p>
                )}
            </div>
            {children && (
                <div className="flex items-center gap-2">
                    {children}
                </div>
            )}
        </div>
    );
}
