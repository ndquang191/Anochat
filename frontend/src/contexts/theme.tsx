"use client";

import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { setSoundEnabled } from "@/hooks/use-sound-notification";

export type Theme = "blue" | "dark" | "pink";

interface ThemeContextType {
	theme: Theme;
	setTheme: (theme: Theme) => void;
	soundEnabled: boolean;
	toggleSound: () => void;
}

const ThemeContext = createContext<ThemeContextType | null>(null);

function applyTheme(theme: Theme) {
	const root = document.documentElement;
	root.classList.remove("dark", "pink");
	if (theme === "dark") root.classList.add("dark");
	else if (theme === "pink") root.classList.add("pink");
}

export function ThemeProvider({ children }: { children: ReactNode }) {
	const [theme, setThemeState] = useState<Theme>("blue");
	const [soundEnabled, setSoundEnabledState] = useState(true);

	useEffect(() => {
		try {
			const savedTheme = localStorage.getItem("theme") as Theme | null;
			if (savedTheme === "dark" || savedTheme === "pink" || savedTheme === "blue") {
				setThemeState(savedTheme);
				applyTheme(savedTheme);
			}
			const savedSound = localStorage.getItem("soundEnabled");
			if (savedSound === "false") {
				setSoundEnabledState(false);
				setSoundEnabled(false);
			}
		} catch {}
	}, []);

	const setTheme = (t: Theme) => {
		setThemeState(t);
		applyTheme(t);
		try { localStorage.setItem("theme", t); } catch {}
	};

	const toggleSound = () => {
		const next = !soundEnabled;
		setSoundEnabledState(next);
		setSoundEnabled(next);
		try { localStorage.setItem("soundEnabled", String(next)); } catch {}
	};

	return (
		<ThemeContext.Provider value={{ theme, setTheme, soundEnabled, toggleSound }}>
			{children}
		</ThemeContext.Provider>
	);
}

export function useTheme() {
	const ctx = useContext(ThemeContext);
	if (!ctx) throw new Error("useTheme must be used within ThemeProvider");
	return ctx;
}
