"use client";

import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { setSoundEnabled } from "@/hooks/use-sound-notification";
import {
	isLanguage,
	LANGUAGE_STORAGE_KEY,
	type Language,
	type TranslationKey,
	translate,
} from "@/lib/i18n";
import { setCookie } from "@/lib/cookies";

export type Theme = "blue" | "dark" | "pink";

interface PreferencesContextType {
	theme: Theme;
	setTheme: (theme: Theme) => void;
	soundEnabled: boolean;
	toggleSound: () => void;
	language: Language;
	setLanguage: (language: Language) => void;
}

const PreferencesContext = createContext<PreferencesContextType | null>(null);

function applyTheme(theme: Theme) {
	const root = document.documentElement;
	root.classList.remove("dark", "pink");
	if (theme === "dark") root.classList.add("dark");
	else if (theme === "pink") root.classList.add("pink");
}

function applyLanguage(language: Language) {
	document.documentElement.lang = language;
}

export function ThemeProvider({
	children,
	initialLanguage,
}: {
	children: ReactNode;
	initialLanguage: Language;
}) {
	const [theme, setThemeState] = useState<Theme>("blue");
	const [soundEnabled, setSoundEnabledState] = useState(true);
	const [language, setLanguageState] = useState<Language>(initialLanguage);

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

	const setTheme = (value: Theme) => {
		setThemeState(value);
		applyTheme(value);
		try {
			localStorage.setItem("theme", value);
		} catch {}
	};

	const setLanguage = (value: Language) => {
		if (!isLanguage(value)) return;
		setLanguageState(value);
		applyLanguage(value);
		setCookie(LANGUAGE_STORAGE_KEY, value, 365);
		try {
			localStorage.setItem(LANGUAGE_STORAGE_KEY, value);
		} catch {}
	};

	const toggleSound = () => {
		const next = !soundEnabled;
		setSoundEnabledState(next);
		setSoundEnabled(next);
		try {
			localStorage.setItem("soundEnabled", String(next));
		} catch {}
	};

	return (
		<PreferencesContext.Provider
			value={{
				theme,
				setTheme,
				soundEnabled,
				toggleSound,
				language,
				setLanguage,
			}}
		>
			{children}
		</PreferencesContext.Provider>
	);
}

function usePreferences() {
	const ctx = useContext(PreferencesContext);
	if (!ctx) throw new Error("useTheme must be used within ThemeProvider");
	return ctx;
}

export function useTheme() {
	const { theme, setTheme, soundEnabled, toggleSound } = usePreferences();
	return { theme, setTheme, soundEnabled, toggleSound };
}

export function useLanguage() {
	const { language, setLanguage } = usePreferences();
	return {
		language,
		setLanguage,
		t: (key: TranslationKey, vars?: Record<string, string | number>) =>
			translate(language, key, vars),
	};
}
