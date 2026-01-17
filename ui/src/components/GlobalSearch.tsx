"use client";

import { useEffect, useState, Suspense } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { Input } from "@/components/ui/input";
import { Search, X } from "lucide-react";
import { Button } from "@/components/ui/button";

function GlobalSearchContent({ placeholder = "Search..." }: { placeholder?: string }) {
    const navigate = useNavigate();
    const location = useLocation();
    const pathname = location.pathname;
    const [searchParams] = useSearchParams();

    const [query, setQuery] = useState(searchParams.get("q") || "");

    useEffect(() => {
        setQuery(searchParams.get("q") || "");
    }, [searchParams]);

    const handleSearch = (value: string) => {
        setQuery(value);
        const params = new URLSearchParams(searchParams.toString());
        if (value) {
            params.set("q", value);
        } else {
            params.delete("q");
        }
        navigate(`${pathname}?${params.toString()}`);
    };

    const clearSearch = () => {
        handleSearch("");
    };

    return (
        <div className="relative w-full md:w-[300px] lg:w-[400px]">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
                placeholder={placeholder}
                value={query}
                onChange={(e) => handleSearch(e.target.value)}
                className="pl-10 pr-10 h-9 bg-muted/50 border-muted-foreground/20 focus-visible:ring-primary/20"
            />
            {query && (
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={clearSearch}
                    className="absolute right-1 top-1/2 -translate-y-1/2 h-7 w-7 hover:bg-transparent text-muted-foreground hover:text-foreground"
                >
                    <X className="h-4 w-4" />
                </Button>
            )}
        </div>
    );
}

export function GlobalSearch({ placeholder }: { placeholder?: string }) {
    return (
        <Suspense fallback={<div className="w-full md:w-[300px] lg:w-[400px] h-9 bg-muted/20 animate-pulse rounded-md" />}>
            <GlobalSearchContent placeholder={placeholder} />
        </Suspense>
    );
}
