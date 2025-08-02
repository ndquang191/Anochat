"use client";

import { useState, useCallback } from "react";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog";

type AlertDialogOptions = {
	title: string;
	description: string;
	confirmText?: string;
	cancelText?: string;
};

export function useAlertDialog() {
	const [isOpen, setIsOpen] = useState(false);
	const [options, setOptions] = useState<AlertDialogOptions | null>(null);
	const [resolve, setResolve] = useState<((value: boolean) => void) | null>(null);

	const open = useCallback((options: AlertDialogOptions) => {
		setOptions(options);
		setIsOpen(true);
		return new Promise<boolean>((res) => {
			setResolve(() => res);
		});
	}, []);

	const handleClose = useCallback(() => {
		setIsOpen(false);
		resolve?.(false);
	}, [resolve]);

	const handleConfirm = useCallback(() => {
		setIsOpen(false);
		resolve?.(true);
	}, [resolve]);

	const dialog = (
		<>
			{options && (
				<AlertDialog open={isOpen}>
					<AlertDialogContent className="h-fit">
						<AlertDialogHeader>
							<AlertDialogTitle>{options.title}</AlertDialogTitle>
							<AlertDialogDescription>{options.description}</AlertDialogDescription>
						</AlertDialogHeader>
						<AlertDialogFooter>
							<AlertDialogCancel onClick={handleClose}>
								{options.cancelText || "Cancel"}
							</AlertDialogCancel>
							<AlertDialogAction onClick={handleConfirm}>
								{options.confirmText || "Confirm"}
							</AlertDialogAction>
						</AlertDialogFooter>
					</AlertDialogContent>
				</AlertDialog>
			)}
		</>
	);

	return {
		dialog,
		open,
	};
}
