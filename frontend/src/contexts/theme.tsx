"use client";

import { createContext, useContext, useEffect, useState, ReactNode } from "react";

export type Theme = "blue" | "dark" | "pink";

interface ThemeContextType {
	theme: Theme;
	setTheme: (theme: Theme) => void;
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

	useEffect(() => {
		try {
			const saved = localStorage.getItem("theme") as Theme | null;
			if (saved === "dark" || saved === "pink" || saved === "blue") {
				setThemeState(saved);
				applyTheme(saved);
			}
		} catch {}
	}, []);

	const setTheme = (t: Theme) => {
		setThemeState(t);
		applyTheme(t);
		try { localStorage.setItem("theme", t); } catch {}
	};

	return (
		<ThemeContext.Provider value={{ theme, setTheme }}>
			{children}
		</ThemeContext.Provider>
	);
}

export function useTheme() {
	const ctx = useContext(ThemeContext);
	if (!ctx) throw new Error("useTheme must be used within ThemeProvider");
	return ctx;
}
