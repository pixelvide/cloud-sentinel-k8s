"use client";

import { useLocation, useNavigate, useSearchParams } from "react-router-dom";

interface NamespaceBadgeProps {
    namespace: string;
    className?: string;
}

export function NamespaceBadge({ namespace, className = "" }: NamespaceBadgeProps) {
    const navigate = useNavigate();
    const location = useLocation();
    const pathname = location.pathname;
    const [searchParams] = useSearchParams();

    const handleClick = () => {
        const params = new URLSearchParams(searchParams.toString());
        params.set("namespace", namespace);
        navigate(`${pathname}?${params.toString()}`);
    };

    return (
        <span
            className={`text-[10px] font-medium bg-secondary text-secondary-foreground px-2 py-0.5 rounded-full whitespace-nowrap cursor-pointer hover:bg-primary hover:text-primary-foreground transition-colors ${className}`}
            onClick={handleClick}
        >
            {namespace}
        </span>
    );
}
