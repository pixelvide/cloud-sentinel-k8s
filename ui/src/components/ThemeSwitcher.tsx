import { useAppearance } from "./appearance-provider";
import { colorThemes, ColorTheme } from "./color-theme-provider";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "./ui/select";
import { Moon, Sun, Palette, Monitor } from "lucide-react";

export function ThemeSwitcher() {
    const { theme, setTheme, colorTheme, setColorTheme } = useAppearance();

    return (
        <div className="flex flex-col gap-4 p-4 border-b bg-muted/20">
            <div className="flex flex-col gap-2">
                <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground opacity-60 flex items-center gap-2">
                    <Monitor className="h-3 w-3" />
                    Interface Mode
                </label>
                <Select value={theme} onValueChange={(v: any) => setTheme(v)}>
                    <SelectTrigger className="w-full bg-background/50 h-9">
                        <SelectValue placeholder="Select mode" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="light">
                            <div className="flex items-center gap-2">
                                <Sun className="h-4 w-4 text-orange-400" />
                                <span>Light</span>
                            </div>
                        </SelectItem>
                        <SelectItem value="dark">
                            <div className="flex items-center gap-2">
                                <Moon className="h-4 w-4 text-primary" />
                                <span>Dark</span>
                            </div>
                        </SelectItem>
                        <SelectItem value="system">
                            <div className="flex items-center gap-2">
                                <Monitor className="h-4 w-4" />
                                <span>System</span>
                            </div>
                        </SelectItem>
                    </SelectContent>
                </Select>
            </div>

            <div className="flex flex-col gap-2">
                <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground opacity-60 flex items-center gap-2">
                    <Palette className="h-3 w-3" />
                    Color Theme
                </label>
                <Select value={colorTheme} onValueChange={(v: ColorTheme) => setColorTheme(v)}>
                    <SelectTrigger className="w-full bg-background/50 h-9">
                        <SelectValue placeholder="Select color" />
                    </SelectTrigger>
                    <SelectContent>
                        {Object.keys(colorThemes).map((themeName) => (
                            <SelectItem key={themeName} value={themeName}>
                                <span className="capitalize">{themeName.replace("-", " ")}</span>
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>
        </div>
    );
}
