"use client";

import { ReactNode } from "react";
import { AuthProvider } from "./auth";
import { AlertDialogProvider } from "./alert-dialog";

interface AppProviderProps {
	children: ReactNode;
}

export function AppProvider({ children }: AppProviderProps) {
	return (
		<AuthProvider>
			<AlertDialogProvider>{children}</AlertDialogProvider>
		</AuthProvider>
	);
}
