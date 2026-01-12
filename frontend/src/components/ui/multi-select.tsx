"use client";

import * as React from "react";
import { ChevronDown, ChevronUp, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
    CommandSeparator,
} from "@/components/ui/command";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";

interface MultiSelectProps {
    options?: {
        label: string;
        value: string;
        icon?: React.ComponentType<{ className?: string }>;
    }[];
    groups?: {
        label: string;
        options: {
            label: string;
            value: string;
            icon?: React.ComponentType<{ className?: string }>;
        }[];
    }[];
    selected: string[];
    onChange: (selected: string[]) => void;
    placeholder?: string;
    loading?: boolean;
    showSearch?: boolean;
    popoverClassName?: string;
}

export function MultiSelect({
    options = [],
    groups = [],
    selected,
    onChange,
    placeholder = "Select...",
    loading = false,
    allOption,
    showSearch = true,
    popoverClassName,
}: MultiSelectProps & { allOption?: { label: string; value: string } }) {
    const [open, setOpen] = React.useState(false);

    const flatOptions = React.useMemo(() => {
        if (groups.length > 0) {
            return groups.flatMap(g => g.options);
        }
        return options;
    }, [groups, options]);

    const handleUnselect = (value: string) => {
        onChange(selected.filter((item) => item !== value));
    };

    const renderOption = (option: { label: string; value: string; icon?: React.ComponentType<{ className?: string }> }) => {
        const isSelected = selected.includes(option.value) || (allOption && selected.includes(allOption.value));
        const toggleOption = () => {
            if (allOption && selected.includes(allOption.value)) {
                // Requirement 1: If "All" is active, clicking changes to Just This One.
                onChange([option.value]);
                return;
            }

            // Is it explicitly in the selected list?
            if (selected.includes(option.value)) {
                onChange(selected.filter((item) => item !== option.value));
            } else {
                const newSelected = [...selected, option.value];
                // Requirement 2: If selecting this makes it "All", switch to "All" value.
                if (allOption && newSelected.length === flatOptions.length) {
                    onChange([allOption.value]);
                } else {
                    onChange(newSelected);
                }
            }
        };

        return (
            <CommandItem
                key={option.value}
                onSelect={toggleOption}
            >
                <Checkbox
                    checked={isSelected}
                    onCheckedChange={toggleOption}
                    className="mr-2"
                />
                {option.icon && (
                    <option.icon className="mr-2 h-4 w-4 text-muted-foreground" />
                )}
                <span>{option.label}</span>
            </CommandItem>
        );
    };

    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <Button
                    variant="outline"
                    role="combobox"
                    aria-expanded={open}
                    className="w-full justify-between hover:bg-muted/50 h-auto min-h-9 py-1 px-3"
                    disabled={loading}
                >
                    <div className="flex gap-1 flex-wrap">
                        {loading ? (
                            <span className="text-muted-foreground">Loading...</span>
                        ) : allOption && selected.includes(allOption.value) ? (
                            <span className="text-foreground">{allOption.label}</span>
                        ) : selected.length === 0 ? (
                            <span className="text-muted-foreground">{placeholder}</span>
                        ) : selected.length === flatOptions.length && flatOptions.length > 0 && !allOption ? (
                            <span className="text-foreground">All Selected ({flatOptions.length})</span>
                        ) : (
                            <div className="flex gap-1 flex-wrap">
                                {selected.length > 2 ? (
                                    <Badge
                                        variant="secondary"
                                        className="rounded-sm px-1 font-normal"
                                    >
                                        {selected.length} selected
                                    </Badge>
                                ) : (
                                    flatOptions
                                        .filter((option) => selected.includes(option.value))
                                        .map((option) => (
                                            <Badge
                                                variant="secondary"
                                                key={option.value}
                                                className="rounded-sm px-1 font-normal mr-1"
                                                onClick={(e: React.MouseEvent) => {
                                                    e.stopPropagation();
                                                    handleUnselect(option.value);
                                                }}
                                            >
                                                {option.label}
                                                <button
                                                    className="ml-1 ring-offset-background rounded-full outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                                                    onKeyDown={(e) => {
                                                        if (e.key === "Enter") {
                                                            handleUnselect(option.value);
                                                        }
                                                    }}
                                                    onMouseDown={(e) => {
                                                        e.preventDefault();
                                                        e.stopPropagation();
                                                    }}
                                                    onClick={(e: React.MouseEvent) => {
                                                        e.preventDefault();
                                                        e.stopPropagation();
                                                        handleUnselect(option.value);
                                                    }}
                                                >
                                                    <X className="h-3 w-3 text-muted-foreground hover:text-foreground" />
                                                </button>
                                            </Badge>
                                        ))
                                )}
                            </div>
                        )}
                    </div>
                    {open ? (
                        <ChevronUp className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                    ) : (
                        <ChevronDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                    )}
                </Button>
            </PopoverTrigger>
            <PopoverContent className={cn("w-[300px] p-0", popoverClassName)} align="start">
                <Command>
                    {showSearch && <CommandInput placeholder="Search..." />}
                    <CommandList>
                        <CommandEmpty>No results found.</CommandEmpty>
                        <CommandGroup>
                            {allOption ? (
                                <>
                                    <CommandItem
                                        onSelect={() => {
                                            if (selected.includes(allOption.value)) {
                                                onChange([]);
                                            } else {
                                                onChange([allOption.value]);
                                            }
                                        }}
                                    >
                                        <Checkbox
                                            checked={selected.includes(allOption.value)}
                                            onCheckedChange={() => {
                                                if (selected.includes(allOption.value)) {
                                                    onChange([]);
                                                } else {
                                                    onChange([allOption.value]);
                                                }
                                            }}
                                            className="mr-2"
                                        />
                                        <span className="font-semibold">{allOption.label}</span>
                                    </CommandItem>
                                    <CommandSeparator className="my-1" />
                                </>
                            ) : (
                                <>
                                    <CommandItem
                                        onSelect={() => {
                                            if (selected.length === flatOptions.length) {
                                                onChange([]);
                                            } else {
                                                onChange(flatOptions.map((option) => option.value));
                                            }
                                        }}
                                    >
                                        <Checkbox
                                            checked={selected.length === flatOptions.length}
                                            onCheckedChange={() => {
                                                if (selected.length === flatOptions.length) {
                                                    onChange([]);
                                                } else {
                                                    onChange(flatOptions.map((option) => option.value));
                                                }
                                            }}
                                            className="mr-2"
                                        />
                                        <span className="font-semibold">Select All</span>
                                    </CommandItem>
                                    <CommandSeparator className="my-1" />
                                </>
                            )}

                            {groups.length > 0 ? (
                                groups.map((group) => (
                                    <React.Fragment key={group.label}>
                                        <div className="px-2 py-1.5 text-xs font-medium text-muted-foreground bg-muted/30">
                                            {group.label}
                                        </div>
                                        {group.options.map(renderOption)}
                                    </React.Fragment>
                                ))
                            ) : (
                                options.map(renderOption)
                            )}
                        </CommandGroup>
                    </CommandList>
                </Command>
            </PopoverContent>
        </Popover>
    );
}
