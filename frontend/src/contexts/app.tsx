"use client";

import { ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/lib/query-client";
import { ErrorBoundary } from "@/components/error-boundary";
import { AuthProvider } from "./auth";
import { AlertDialogProvider } from "./alert-dialog";
import { ThemeProvider } from "./theme";
import type { Language } from "@/lib/i18n";

interface AppProviderProps {
	children: ReactNode;
	initialLanguage: Language;
}

export function AppProvider({ children, initialLanguage }: AppProviderProps) {
	return (
		<ThemeProvider initialLanguage={initialLanguage}>
			<QueryClientProvider client={queryClient}>
				<ErrorBoundary>
					<AuthProvider>
						<AlertDialogProvider>{children}</AlertDialogProvider>
					</AuthProvider>
				</ErrorBoundary>
			</QueryClientProvider>
		</ThemeProvider>
	);
}
