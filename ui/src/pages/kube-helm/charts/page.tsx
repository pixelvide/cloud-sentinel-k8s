"use client";

export default function HelmChartsPage() {
    return (
        <div className="flex flex-col h-full bg-background/50 backdrop-blur-sm">
            {/* PageHeader removed as per user request */}

            <div className="flex-1 p-6 flex items-center justify-center">
                <div className="text-center space-y-4 max-w-md mx-auto p-12 rounded-2xl border border-dashed border-border bg-card/30">
                    <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                        <span className="text-2xl">⚓️</span>
                    </div>
                    <h3 className="text-lg font-semibold text-foreground">Coming Soon</h3>
                    <p className="text-muted-foreground text-sm leading-relaxed">
                        The Helm Charts repository browser and installer is currently under development. Check back soon
                        for updates!
                    </p>
                </div>
            </div>
        </div>
    );
}
