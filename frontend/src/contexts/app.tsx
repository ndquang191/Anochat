"use client";

import { ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/lib/query-client";
import { ErrorBoundary } from "@/components/error-boundary";
import { AuthProvider } from "./auth";
import { AlertDialogProvider } from "./alert-dialog";

interface AppProviderProps {
	children: ReactNode;
}

export function AppProvider({ children }: AppProviderProps) {
	return (
		<QueryClientProvider client={queryClient}>
			<ErrorBoundary>
				<AuthProvider>
					<AlertDialogProvider>{children}</AlertDialogProvider>
				</AuthProvider>
			</ErrorBoundary>
		</QueryClientProvider>
	);
}
