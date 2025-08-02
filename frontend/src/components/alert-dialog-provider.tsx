"use client";

import { createContext, useContext } from "react";
import { useAlertDialog } from "@/hooks/use-alert-dialog";

const AlertDialogContext = createContext<ReturnType<typeof useAlertDialog> | null>(null);

export function AlertDialogProvider({ children }: { children: React.ReactNode }) {
	const alertDialog = useAlertDialog();

	return (
		<AlertDialogContext.Provider value={alertDialog}>
			{alertDialog.dialog}
			{children}
		</AlertDialogContext.Provider>
	);
}

export function useAlertDialogContext() {
	const context = useContext(AlertDialogContext);
	if (!context) {
		throw new Error("useAlertDialogContext must be used within AlertDialogProvider");
	}
	return context;
}
