"use client";

import * as React from "react";
import { ChevronsUpDown, X } from "lucide-react";
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
    options: {
        label: string;
        value: string;
        icon?: React.ComponentType<{ className?: string }>;
    }[];
    selected: string[];
    onChange: (selected: string[]) => void;
    placeholder?: string;
    loading?: boolean;
}

export function MultiSelect({
    options,
    selected,
    onChange,
    placeholder = "Select...",
    loading = false,
}: MultiSelectProps) {
    const [open, setOpen] = React.useState(false);

    const handleUnselect = (value: string) => {
        onChange(selected.filter((item) => item !== value));
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
                        ) : selected.length === 0 ? (
                            <span className="text-muted-foreground">{placeholder}</span>
                        ) : selected.length === options.length && options.length > 0 ? (
                            <span className="text-foreground">All Selected ({options.length})</span>
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
                                    options
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
                    <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-[300px] p-0" align="start">
                <Command>
                    <CommandInput placeholder="Search..." />
                    <CommandList>
                        <CommandEmpty>No results found.</CommandEmpty>
                        <CommandGroup>
                            <CommandItem
                                onSelect={() => {
                                    if (selected.length === options.length) {
                                        onChange([]);
                                    } else {
                                        onChange(options.map((option) => option.value));
                                    }
                                }}
                            >
                                <Checkbox
                                    checked={selected.length === options.length}
                                    onCheckedChange={() => {
                                        if (selected.length === options.length) {
                                            onChange([]);
                                        } else {
                                            onChange(options.map((option) => option.value));
                                        }
                                    }}
                                    className="mr-2"
                                />
                                <span className="font-semibold">Select All</span>
                            </CommandItem>
                            <CommandSeparator className="my-1" />
                            {options.map((option) => {
                                const isSelected = selected.includes(option.value);
                                return (
                                    <CommandItem
                                        key={option.value}
                                        onSelect={() => {
                                            const allSelected = selected.length === options.length;
                                            if (isSelected) {
                                                // If all are selected and user clicks one, select only that one
                                                if (allSelected) {
                                                    onChange([option.value]);
                                                } else {
                                                    onChange(selected.filter((item) => item !== option.value));
                                                }
                                            } else {
                                                onChange([...selected, option.value]);
                                            }
                                        }}
                                    >
                                        <Checkbox
                                            checked={isSelected}
                                            onCheckedChange={() => {
                                                const allSelected = selected.length === options.length;
                                                if (isSelected) {
                                                    if (allSelected) {
                                                        onChange([option.value]);
                                                    } else {
                                                        onChange(selected.filter((item) => item !== option.value));
                                                    }
                                                } else {
                                                    onChange([...selected, option.value]);
                                                }
                                            }}
                                            className="mr-2"
                                        />
                                        {option.icon && (
                                            <option.icon className="mr-2 h-4 w-4 text-muted-foreground" />
                                        )}
                                        <span>{option.label}</span>
                                    </CommandItem>
                                );
                            })}
                        </CommandGroup>
                    </CommandList>
                </Command>
            </PopoverContent>
        </Popover>
    );
}
